package wiring

import (
	"time"

	"github.com/google/uuid"
)

// Clock implements a real-time clock for production use.
type Clock struct{}

// Now returns the current UTC time.
func (Clock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates real UUID strings for production use.
type UUIDGen struct{}

// New returns a new random UUID string.
func (UUIDGen) New() string { return uuid.NewString() }
