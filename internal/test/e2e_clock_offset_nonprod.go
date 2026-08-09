//go:build !production

package test

import (
	"sync/atomic"
	"time"
)

var e2eClockOffsetSeconds int64

// SetE2EClockOffset sets the e2e-test clock offset (in seconds) applied on
// top of the real wall clock. Non-production builds only.
func SetE2EClockOffset(seconds int64) {
	atomic.StoreInt64(&e2eClockOffsetSeconds, seconds)
}

// AdvanceE2EClock adds seconds to the current e2e-test clock offset.
func AdvanceE2EClock(seconds int64) {
	atomic.AddInt64(&e2eClockOffsetSeconds, seconds)
}

// ResetE2EClockOffset clears the e2e-test clock offset back to zero.
func ResetE2EClockOffset() {
	SetE2EClockOffset(0)
}

// E2EClockOffset returns the current e2e-test clock offset as a Duration, to
// be added to time.Now() by callers that honor the e2e clock.
func E2EClockOffset() time.Duration {
	return time.Duration(atomic.LoadInt64(&e2eClockOffsetSeconds)) * time.Second
}
