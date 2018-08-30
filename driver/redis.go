package driver

import (
	"fmt"
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
	return redis.String(c.Do("GET", key))
}

// Set key-value pair
func (r *redisDriver) Set(key string, value string) error {
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
	arr, err := redis.Strings(c.Do("MGET", tmp...))
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
func (r *redisDriver) MSet(kvs map[string]string) error {
	return nil
}

// Del delete specified key
func (r *redisDriver) Del(key string) error {
	return nil
}

// Check if the given key exists
func (r *redisDriver) Exists(key string) (bool, error) {
	return false, nil
}

// Expire set key expiration
func (r *redisDriver) Expire(key string, ex int64) error {
	return nil
}

// Incr increment key
func (r *redisDriver) Incr(key string, delta string) (string, error) {
	return "", nil
}

// Decr increment key
func (r *redisDriver) Decr(key string, delta string) (string, error) {
	return "", nil
}

// func for hashes

// HGEt get hash key
func (r *redisDriver) HGet(key string, hk string) (string, error) {
	return "", nil
}

// HSet set hash key
func (r *redisDriver) HSet(key string, hk string, value string) error {
	return nil
}

// HMGet get multiple hash keys
func (r *redisDriver) HMGet(key string, hks []string) (map[string]string, error) {
	return nil, nil
}

// HMSet set multiple hash keys
func (r *redisDriver) HMSet(key string, kvs map[string]string) error {
	return nil
}

// HGetAll get all hash keys
func (r *redisDriver) HGetAll(key string) (map[string]string, error) {
	return nil, nil
}

// HDel delete hash key
func (r *redisDriver) HDel(key string, hk string) error {
	return nil
}

// HExists check if the given hash key exists
func (r *redisDriver) HExists(key string, hk string) (bool, error) {
	return false, nil
}

// HIncr increment value of hash key
func (r *redisDriver) HIncr(key string, hk string, delta string) (string, error) {
	return "", nil
}

// HDecr decrement value of hash key
func (r *redisDriver) HDecr(key string, hk string, delta string) (string, error) {
	return "", nil
}
