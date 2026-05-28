package bootstrap

import (
	"testing"
	"time"

	"metaldocs/internal/platform/config"
)

func TestWorkerClaimLease_UsesMinimumFloor(t *testing.T) {
	got := workerClaimLease(config.WorkerConfig{RetryMaxSeconds: 30})
	if got != 5*time.Minute {
		t.Fatalf("workerClaimLease() = %v, want %v", got, 5*time.Minute)
	}
}

func TestWorkerClaimLease_UsesRetryMaxWhenLonger(t *testing.T) {
	got := workerClaimLease(config.WorkerConfig{RetryMaxSeconds: 900})
	if got != 15*time.Minute {
		t.Fatalf("workerClaimLease() = %v, want %v", got, 15*time.Minute)
	}
}
