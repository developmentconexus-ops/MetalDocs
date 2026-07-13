package contracts

import (
	"errors"
	"fmt"
)

// ErrValidation is the sentinel wrapped by every contract Validate failure, so
// callers can errors.Is against it regardless of the specific field that failed.
var ErrValidation = errors.New("approval contract validation failed")

// wrapValidation wraps err with ErrValidation, or returns nil if err is nil.
func wrapValidation(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrValidation, err)
}
