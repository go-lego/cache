package driver

import (
	"errors"
	"fmt"
	"strings"
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
	c := r.pool.Get()
	defer c.Close()
	tmp := make([]interface{}, len(kvs)*2)
	i := 0
	for k, v := range kvs {
		tmp[i] = k
		tmp[i+1] = v
		i += 2
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
	_, err := redis.String(c.Do("DEL", key))
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
	_, err := redis.String(c.Do("EXPIRE", key, ex))
	return err
}

// Incr increment key
func (r *redisDriver) Incr(key string, delta string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.String(c.Do("INCRBY", key, delta))
}

// Decr increment key
func (r *redisDriver) Decr(key string, delta string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.String(c.Do("DECRBY", key, delta))
}

// func for hashes

// HGEt get hash key
func (r *redisDriver) HGet(key string, hk string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.String(c.Do("HGET", key, hk))
}

// HSet set hash key
func (r *redisDriver) HSet(key string, hk string, value string) error {
	c := r.pool.Get()
	defer c.Close()
	_, err := redis.String(c.Do("HSET", key, hk, value))
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
	return redis.StringMap(c.Do("HMGET", tmp...))
}

// HMSet set multiple hash keys
func (r *redisDriver) HMSet(key string, kvs map[string]string) error {
	c := r.pool.Get()
	defer c.Close()
	tmp := make([]interface{}, len(kvs)*2+1)
	tmp[0] = key
	i := 1
	for k, v := range kvs {
		tmp[i] = k
		tmp[i+1] = v
		i += 2
	}
	_, err := redis.String(c.Do("HMSET", tmp...))
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
	_, err := redis.String(c.Do("HDEL", key, hk))
	return err
}

// HExists check if the given hash key exists
func (r *redisDriver) HExists(key string, hk string) (bool, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.Bool(c.Do("HEXISTS", key, hk))
}

// HIncr increment value of hash key
func (r *redisDriver) HIncr(key string, hk string, delta string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	return redis.String(c.Do("HINCRBY", key, hk, delta))
}

// HDecr decrement value of hash key
func (r *redisDriver) HDecr(key string, hk string, delta string) (string, error) {
	c := r.pool.Get()
	defer c.Close()
	nd := ""
	i := strings.Index(delta, "-")
	if i == -1 {
		nd = fmt.Sprintf("-%s", delta)
	} else if i == 0 {
		nd = delta[1:]
	} else {
		return "", errors.New("driver redis: invalid delta string")
	}
	return redis.String(c.Do("HINCRBY", key, hk, nd))
}
