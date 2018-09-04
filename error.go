package cache

import (
	"errors"
	"fmt"
)

var (
	// ErrValueNil cache value is nil, indicates that key not exist
	ErrValueNil = errors.New("cache: value nil")
)

// InternalError generate interfanl error
func InternalError(err error) error {
	return fmt.Errorf("cache: internal error (%s)", err)
}
