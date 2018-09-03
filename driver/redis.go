package driver

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
)

type redisPool interface {
	Get() redis.Conn
}

// redisDriver Redis cache driver implementation
type redisDriver struct {
	options Options
	pool    redisPool
	test    bool // test mode is used for fixing the issue caused by map iterating
}

// newredisDriver create new redis cache
func newRedisDriver(opts Options) Driver {
	c := &redisDriver{
		options: opts,
	}
	return c
}

// Options get options
func (r *redisDriver) Options() Options {
	return r.options
}

// Init initialize redis connection
func (r *redisDriver) Init() error {
	opts := r.options
	p := &redis.Pool{
		MaxIdle:     5,
		MaxActive:   0,
		IdleTimeout: 2 * time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%d", opts.Host, opts.Port), redis.DialPassword(opts.Password))
		},
	}
	t := p.Get()
	defer t.Close()
	if _, err := t.Do("PING"); err != nil {
		return err
	}
	r.pool = p
	return nil
}

// func for keys

// Get value by key
func (r *redisDriver) Get(key string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	v, err := redis.String(c.Do("GET", key))
	if err != nil && err == redis.ErrNil {
		return "", ErrValueNil
	}
	return v, err
}

// Set key-value pair
func (r *redisDriver) Set(key string, value interface{}) error {
	c := r.pool.Get()
	defer c.Close()
	_, err := c.Do("SET", key, value)
	return err
}

// MGet get multiple keys
func (r *redisDriver) MGet(keys []string) (map[string]string, error) {
	c := r.pool.Get()
	defer c.Close()
	tmp := make([]interface{}, len(keys))
	for i, k := range keys {
		tmp[i] = k
	}
	// fmt.Println("Arguments:", tmp)
	arr, err := redis.Strings(c.Do("MGET", tmp...))
	// fmt.Println("Response:", arr)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string, len(keys))
	for i, k := range keys {
		ret[k] = arr[i]
	}
	return ret, nil
}

// MSet set multiple key-value pairs
func (r *redisDriver) MSet(kvs map[string]interface{}) error {
	c := r.pool.Get()
	defer c.Close()
	tmp := make([]interface{}, len(kvs)*2)
	i := 0
	if r.test { // sort keys in test mode
		sks := []string{}
		for k := range kvs {
			sks = append(sks, k)
		}
		sort.Slice(sks, func(i, j int) bool { return sks[i] < sks[j] })
		for _, k := range sks {
			tmp[i] = k
			tmp[i+1] = kvs[k]
			i += 2
		}
	} else {
		for k, v := range kvs {
			tmp[i] = k
			tmp[i+1] = v
			i += 2
		}
	}
	_, err := redis.String(c.Do("MSET", tmp...))
	if err != nil {
		return err
	}
	return nil
}

// Del delete specified key
func (r *redisDriver) Del(key string) error {
	c := r.pool.Get()
	defer c.Close()
	_, err := c.Do("DEL", key)
	return err
}

// Check if the given key exists
func (r *redisDriver) Exists(key string) (bool, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.Bool(c.Do("EXISTS", key))
}

// Expire set key expiration
func (r *redisDriver) Expire(key string, ex int64) error {
	c := r.pool.Get()
	defer c.Close()
	_, err := c.Do("EXPIRE", key, ex)
	return err
}

// Incr increment key
func (r *redisDriver) Incr(key string, delta interface{}) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	switch delta.(type) {
	case int, int32, int64:
		v, err := redis.Int64(c.Do("INCRBY", key, delta))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return redis.String(c.Do("INCRBYFLOAT", key, delta))
	}
	return "", errors.New("driver redis: invalid delta value")
}

// Decr increment key
func (r *redisDriver) Decr(key string, delta interface{}) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	switch delta.(type) {
	case int, int32, int64:
		v, err := redis.Int64(c.Do("DECRBY", key, delta))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	case float32:
		return redis.String(c.Do("INCRBYFLOAT", key, -delta.(float32)))
	case float64:
		return redis.String(c.Do("INCRBYFLOAT", key, -delta.(float64)))
	}
	return "", errors.New("driver redis: invalid delta value")
}

// func for hashes

// HGEt get hash key
func (r *redisDriver) HGet(key string, hk string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.String(c.Do("HGET", key, hk))
}

// HSet set hash key
func (r *redisDriver) HSet(key string, hk string, value interface{}) error {
	c := r.pool.Get()
	defer c.Close()
	_, err := c.Do("HSET", key, hk, value)
	return err
}

// HMGet get multiple hash keys
func (r *redisDriver) HMGet(key string, hks []string) (map[string]string, error) {
	c := r.pool.Get()
	defer c.Close()
	tmp := make([]interface{}, len(hks)+1)
	tmp[0] = key
	for i, k := range hks {
		tmp[i+1] = k
	}
	vals, err := redis.Strings(c.Do("HMGET", tmp...))
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string, len(key)*2)
	for i, v := range vals {
		ret[hks[i]] = v
	}
	return ret, nil
}

// HMSet set multiple hash keys
func (r *redisDriver) HMSet(key string, kvs map[string]interface{}) error {
	c := r.pool.Get()
	defer c.Close()
	tmp := make([]interface{}, len(kvs)*2+1)
	tmp[0] = key
	i := 1
	if r.test { // sort keys in test mode
		sks := []string{}
		for k := range kvs {
			sks = append(sks, k)
		}
		sort.Slice(sks, func(i, j int) bool { return sks[i] < sks[j] })
		for _, k := range sks {
			tmp[i] = k
			tmp[i+1] = kvs[k]
			i += 2
		}
	} else {
		for k, v := range kvs {
			tmp[i] = k
			tmp[i+1] = v
			i += 2
		}
	}
	_, err := c.Do("HMSET", tmp...)
	return err
}

// HGetAll get all hash keys
func (r *redisDriver) HGetAll(key string) (map[string]string, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.StringMap(c.Do("HGETALL", key))
}

// HDel delete hash key
func (r *redisDriver) HDel(key string, hk string) error {
	c := r.pool.Get()
	defer c.Close()
	_, err := c.Do("HDEL", key, hk)
	return err
}

// HExists check if the given hash key exists
func (r *redisDriver) HExists(key string, hk string) (bool, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.Bool(c.Do("HEXISTS", key, hk))
}

// HIncr increment value of hash key
func (r *redisDriver) HIncr(key string, hk string, delta interface{}) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	switch delta.(type) {
	case int, int32, int64:
		v, err := redis.Int64(c.Do("HINCRBY", key, hk, delta))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return redis.String(c.Do("HINCRBYFLOAT", key, hk, delta))
	}
	return "", errors.New("driver redis: invalid delta value")
}

// HDecr decrement value of hash key
func (r *redisDriver) HDecr(key string, hk string, delta interface{}) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	switch delta.(type) {
	case int:
		v, err := redis.Int64(c.Do("HINCRBY", key, hk, -delta.(int)))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	case int32:
		v, err := redis.Int64(c.Do("HINCRBY", key, hk, -delta.(int32)))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	case int64:
		v, err := redis.Int64(c.Do("HINCRBY", key, hk, -delta.(int64)))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d", v), nil
	case float32:
		return redis.String(c.Do("HINCRBYFLOAT", key, hk, -delta.(float32)))
	case float64:
		return redis.String(c.Do("HINCRBYFLOAT", key, hk, -delta.(float64)))
	}
	return "", errors.New("driver redis: invalid delta value")
}

// BeforeCreate called before transaction creation
func (r *redisDriver) BeforeCreate() error {
	return nil
}

// AfterCreate called after transaction creation
func (r *redisDriver) AfterCreate() error {
	return nil
}

// BeforeCommit called before transaction commit
func (r *redisDriver) BeforeCommit() error {
	return nil
}

// AfterCommit called after transaction commit
func (r *redisDriver) AfterCommit() error {
	return nil
}

// BeforeRollback called before transaction rollback
func (r *redisDriver) BeforeRollback() error {
	return nil
}

// AfterRollback called after transaction rollback
func (r *redisDriver) AfterRollback() error {
	return nil
}
