package driver

// Options for driver
type Options struct {
	Type     string // current supported type is: "redis"
	Host     string // host of server
	Port     int    // port of server
	Password string // password if needed
}

// Option dynamic option func
type Option func(*Options)

// newOptions create new options
func newOptions(opts ...Option) Options {
	opt := Options{
		Type:     "redis",
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "",
	}

	for _, o := range opts {
		o(&opt)
	}
	return opt
}

// Type option
func Type(t string) Option {
	return func(opts *Options) {
		opts.Type = t
	}
}

// Host option
func Host(h string) Option {
	return func(opts *Options) {
		opts.Host = h
	}
}

// Port option
func Port(p int) Option {
	return func(opts *Options) {
		opts.Port = p
	}
}

// Password option
func Password(p string) Option {
	return func(opts *Options) {
		opts.Password = p
	}
}
