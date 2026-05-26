package contracts

import (
	"errors"
	"fmt"
)

var ErrValidation = errors.New("approval contract validation failed")

func wrapValidation(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrValidation, err)
}
