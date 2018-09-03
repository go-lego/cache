package driver

import "errors"

// Driver cache driver interface
// Use mockgen to generate mocked struct:
// mockgen -package mock -destination driver/mock/driver.mock.go github.com/go-lego/cache/driver Driver
type Driver interface {
	// Options get options
	Options() Options

	// Init initialize
	Init() error

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

var (
	// ErrNotSupported error indicates that the cache type is not supported yet
	ErrNotSupported = errors.New("Cache type not supported yet")

	// ErrValueNil cache value nil, indicates that key not exist
	ErrValueNil = errors.New("driver: value nil")

	// DefaultDriver default driver
	DefaultDriver = newRedisDriver(newOptions())
)

// NewDriver create new cache instance
func NewDriver(opts ...Option) Driver {
	options := newOptions(opts...)
	if options.Type == "redis" {
		return newRedisDriver(options)
	}
	return nil
}
