package cache

import (
	"fmt"

	"github.com/go-lego/cache/driver"
)

// Cache interface
type Cache interface {
	// Options get options
	Options() Options

	// Init initialize
	Init() error

	// FlushMemory flush data in memory
	FlushMemory()

	// BeginTransaction start a transaction if none active
	BeginTransaction() Transaction

	// func for keys

	// Get value by key
	Get(key string) (string, error)

	// Set key-value pair
	Set(key string, value interface{}) error

	// MGet get multiple keys
	MGet(keys []string) (map[string]string, error)

	// MSet set multiple key-value pairs
	MSet(kvs map[string]interface{}) error

	// Del delete specified key
	Del(key string) error

	// Check if the given key exists
	Exists(key string) (bool, error)

	// Expire set key expiration
	Expire(key string, ex int64) error

	// Incr increment key
	Incr(key string, delta interface{}) (string, error)

	// Decr Decrement key
	Decr(key string, delta interface{}) (string, error)

	// func for hashes

	// HGEt get hash key
	HGet(key string, hk string) (string, error)

	// HSet set hash key
	HSet(key string, hk string, value interface{}) error

	// HMGet get multiple hash keys
	HMGet(key string, hks []string) (map[string]string, error)

	// HMSet set multiple hash keys
	HMSet(key string, kvs map[string]interface{}) error

	// HGetAll get all hash keys
	HGetAll(key string) (map[string]string, error)

	// HDel delete hash key
	HDel(key string, hk string) error

	// HExists check if the given hash key exists
	HExists(key string, hk string) (bool, error)

	// HIncr increment value of hash key
	HIncr(key string, hk string, delta interface{}) (string, error)

	// HDecr decrement value of hash key
	HDecr(key string, hk string, delta interface{}) (string, error)
}

// NewCache create new cache instance
func NewCache(opts ...Option) Cache {
	return newCacheImpl(opts...)
}

const flagValueNil = "__value_nil__"

// cacheImpl cache implementation
type cacheImpl struct {
	options Options
	tx      *transImpl

	keys    map[string]string
	hsets   map[string]map[string]string
	delKeys map[string]string
}

// newCacheImpl create new cacheImpl
func newCacheImpl(opts ...Option) *cacheImpl {
	options := newOptions(opts...)
	c := &cacheImpl{
		options: options,
		keys:    make(map[string]string),
		hsets:   make(map[string]map[string]string),
		delKeys: make(map[string]string),
	}
	if c.options.Driver == nil {
		c.options.Driver = driver.DefaultDriver
	}
	return c
}

// Options get options
func (c *cacheImpl) Options() Options {
	return c.options
}

// Init initialize cache
func (c *cacheImpl) Init() error {
	return c.options.Driver.Init()
}

// FlushMemory flush data in memory
func (c *cacheImpl) FlushMemory() {
	c.keys = make(map[string]string)
	c.delKeys = make(map[string]string)
	c.hsets = make(map[string]map[string]string)
}

// BeginTransaction start a transaction if none active
func (c *cacheImpl) BeginTransaction() Transaction {
	if c.tx == nil || c.tx.active == false {
		c.tx = newTransImpl(c)
	}
	return c.tx
}

// getCurrentTransaction get current active transaction
func (c *cacheImpl) getCurrentTransaction() *transImpl {
	if c.tx != nil && !c.tx.active {
		c.tx = nil
	}
	return c.tx
}

// func for keys

// Get value by key
func (c *cacheImpl) Get(key string) (string, error) {
	if _, ok := c.delKeys[key]; ok { // already deleted
		return "", ErrValueNil
	}
	if v, ok := c.keys[key]; ok {
		if v == flagValueNil { // get before but not found
			return "", ErrValueNil
		}
		return v, nil
	}
	v, err := c.options.Driver.Get(key)
	if err == nil {
		c.keys[key] = v
	} else if err == driver.ErrValueNil { // not found, set nil value flag
		c.keys[key] = flagValueNil
	}
	return v, err
}

// ValueToString convert interface{} value to string
func ValueToString(value interface{}) string {
	switch value.(type) {
	case int, int32, int64:
		return fmt.Sprintf("%d", value)
	case float32, float64:
		return fmt.Sprintf("%f", value)
	case string:
		return value.(string)
	case []byte:
		return string(value.([]byte))
	}
	return ""
}

// Set key-value pair
func (c *cacheImpl) Set(key string, value interface{}) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onSet(key, value)
		delete(c.delKeys, key)
		c.keys[key] = ValueToString(value)
		return nil
	}
	err := c.options.Driver.Set(key, value)
	if err == nil {
		c.keys[key] = ValueToString(value)
	}
	return err
}

// MGet get multiple keys
func (c *cacheImpl) MGet(keys []string) (map[string]string, error) {
	hits := map[string]string{}
	noh := []string{}
	for _, k := range keys {
		if _, ok := c.delKeys[k]; ok {
			hits[k] = ""
			continue
		}
		if v, ok := c.keys[k]; ok {
			if v == flagValueNil {
				hits[k] = ""
			} else {
				hits[k] = v
			}
			continue
		}
		noh = append(noh, k)
	}
	if len(noh) > 0 {
		nm, err := c.options.Driver.MGet(noh)
		if err != nil {
			return nil, err
		}
		for k, v := range nm { // merge new map to hits
			hits[k] = v
		}
	}
	return hits, nil
}

// MSet set multiple key-value pairs
func (c *cacheImpl) MSet(kvs map[string]interface{}) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onMSet(kvs)
		for k, v := range kvs {
			delete(c.delKeys, k)
			c.keys[k] = ValueToString(v)
		}
		return nil
	}
	err := c.options.Driver.MSet(kvs)
	if err == nil {
		for k, v := range kvs {
			delete(c.delKeys, k)
			c.keys[k] = ValueToString(v)
		}
	}
	return err
}

// Del delete specified key
func (c *cacheImpl) Del(key string) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onDel(key)
		delete(c.keys, key)
		c.delKeys[key] = ""
		return nil
	}
	err := c.options.Driver.Del(key)
	if err == nil {
		delete(c.keys, key)
		c.delKeys[key] = ""
	}
	return err
}

// Check if the given key exists
func (c *cacheImpl) Exists(key string) (bool, error) {
	if _, ok := c.delKeys[key]; ok { // already deleted
		return false, nil
	}
	if v, ok := c.keys[key]; ok { // already loaded into memory
		if v != flagValueNil {
			return true, nil
		}
		return false, nil
	}
	if _, ok := c.hsets[key]; ok { // already loaded into memory
		return true, nil
	}
	return c.options.Driver.Exists(key)
}

// Expire set key expiration
func (c *cacheImpl) Expire(key string, ex int64) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onExpire(key, ex)
		return nil
	}
	return c.options.Driver.Expire(key, ex)
}

// Incr increment key
func (c *cacheImpl) Incr(key string, delta interface{}) (string, error) {
	nv, err := c.options.Driver.Incr(key, delta)
	if err != nil {
		return "", err
	}
	delete(c.delKeys, key)
	c.keys[key] = nv
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onIncr(key, delta)
	}
	return nv, err
}

// Decr increment key
func (c *cacheImpl) Decr(key string, delta interface{}) (string, error) {
	nv, err := c.options.Driver.Decr(key, delta)
	if err != nil {
		return "", err
	}
	delete(c.delKeys, key)
	c.keys[key] = nv
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onDecr(key, delta)
	}
	return nv, err
}

// func for hashes

func (c *cacheImpl) setMemoryHashSet(key string, hk string, val string) {
	m, ok := c.hsets[key]
	if !ok {
		m = map[string]string{}
	}
	m[hk] = val
	c.hsets[key] = m
}

// HGEt get hash key
func (c *cacheImpl) HGet(key string, hk string) (string, error) {
	if _, ok := c.delKeys[key]; ok { // key is deleted
		return "", ErrValueNil
	}
	if m, ok := c.hsets[key]; ok { // hash set is loaded into memory
		if v, o := m[hk]; o {
			if v == flagValueNil {
				return "", ErrValueNil
			}
			return v, nil
		}
	}
	v, err := c.options.Driver.HGet(key, hk)
	if err != nil && err != driver.ErrValueNil {
		return "", err
	}
	if err == driver.ErrValueNil {
		v = flagValueNil
		err = ErrValueNil
	}
	delete(c.delKeys, key)
	c.setMemoryHashSet(key, hk, v)
	return v, err
}

// HSet set hash key
func (c *cacheImpl) HSet(key string, hk string, value interface{}) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onHSet(key, hk, value)
		delete(c.delKeys, key)
		c.setMemoryHashSet(key, hk, ValueToString(value))
		return nil
	}
	err := c.options.Driver.HSet(key, hk, value)
	if err == nil {
		delete(c.delKeys, key)
		c.setMemoryHashSet(key, hk, ValueToString(value))
	}
	return err
}

// HMGet get multiple hash keys
func (c *cacheImpl) HMGet(key string, hks []string) (map[string]string, error) {
	hits := map[string]string{}
	if _, ok := c.delKeys[key]; ok { // already deleted whole key
		return hits, nil
	}
	noh := []string{}
	if m, ok := c.hsets[key]; ok {
		for _, hk := range hks {
			if v, o := m[hk]; o {
				if v == flagValueNil {
					hits[hk] = ""
				} else {
					hits[hk] = v
				}
			} else {
				noh = append(noh, hk)
			}
		}
	} else {
		noh = hks
	}

	if len(noh) > 0 {
		nm, err := c.options.Driver.HMGet(key, noh)
		if err != nil {
			return nil, err
		}
		for k, v := range nm {
			hits[k] = v
			c.setMemoryHashSet(key, k, v)
		}
	}

	return hits, nil
}

// HMSet set multiple hash keys
func (c *cacheImpl) HMSet(key string, kvs map[string]interface{}) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onHMSet(key, kvs)
		delete(c.delKeys, key)
		for k, v := range kvs {
			c.setMemoryHashSet(key, k, ValueToString(v))
		}
		return nil
	}

	err := c.options.Driver.HMSet(key, kvs)
	if err == nil {
		delete(c.delKeys, key)
		for k, v := range kvs {
			c.setMemoryHashSet(key, k, ValueToString(v))
		}
	}
	return err
}

// HGetAll get all hash keys
func (c *cacheImpl) HGetAll(key string) (map[string]string, error) {
	if _, ok := c.delKeys[key]; ok {
		return map[string]string{}, nil
	}
	ret, err := c.options.Driver.HGetAll(key)
	if err != nil {
		return nil, err
	}
	if m, ok := c.hsets[key]; ok {
		for k, v := range m {
			if v == flagValueNil {
				ret[k] = ""
			} else {
				ret[k] = v
			}
		}
	}
	return ret, err
}

// HDel delete hash key
func (c *cacheImpl) HDel(key string, hk string) error {
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onHDel(key, hk)
		c.setMemoryHashSet(key, hk, flagValueNil)
		return nil
	}

	err := c.options.Driver.HDel(key, hk)
	if err == nil {
		c.setMemoryHashSet(key, hk, flagValueNil)
	}
	return err
}

// HExists check if the given hash key exists
func (c *cacheImpl) HExists(key string, hk string) (bool, error) {
	if _, ok := c.delKeys[key]; ok {
		return false, nil
	}
	if m, ok := c.hsets[key]; ok {
		if v, o := m[hk]; o {
			if v == flagValueNil {
				return false, nil
			}
			return true, nil
		}
	}
	return c.options.Driver.HExists(key, hk)
}

// HIncr increment value of hash key
func (c *cacheImpl) HIncr(key string, hk string, delta interface{}) (string, error) {
	nv, err := c.options.Driver.HIncr(key, hk, delta)
	if err != nil {
		return "", err
	}
	delete(c.delKeys, key)
	c.setMemoryHashSet(key, hk, nv)
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onHIncr(key, hk, delta)
	}
	return nv, err
}

// HDecr decrement value of hash key
func (c *cacheImpl) HDecr(key string, hk string, delta interface{}) (string, error) {
	nv, err := c.options.Driver.HDecr(key, hk, delta)
	if err != nil {
		return "", err
	}
	delete(c.delKeys, key)
	c.setMemoryHashSet(key, hk, nv)
	tx := c.getCurrentTransaction()
	if tx != nil {
		tx.onHDecr(key, hk, delta)
	}
	return nv, err
}
