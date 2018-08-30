package cache

import "github.com/go-lego/cache/driver"

// Options for cache
type Options struct {
	Driver driver.Driver
}

// Option func
type Option func(*Options)

// newOptions create new options
func newOptions(opts ...Option) Options {
	opt := Options{}

	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// Driver option
func Driver(d driver.Driver) Option {
	return func(opts *Options) {
		opts.Driver = d
	}
}
