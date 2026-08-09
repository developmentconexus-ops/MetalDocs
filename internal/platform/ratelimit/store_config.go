package ratelimit

import (
	"fmt"
	"strconv"
	"strings"
)

// Env var names for store backend selection (M8/F8.2 Task A). These are the
// only rate-limit env knobs today — per-route quotas remain compiled-in via
// NewConfig/DefaultConfig (see the RouteKey doc comment in config.go).
const (
	EnvRatelimitStore         = "METALDOCS_RATELIMIT_STORE"
	EnvRatelimitRedisAddr     = "METALDOCS_RATELIMIT_REDIS_ADDR"
	EnvRatelimitRedisPassword = "METALDOCS_RATELIMIT_REDIS_PASSWORD" // #nosec G101 -- this is the env VAR NAME (the lookup key), not a credential value; the actual password is read from the environment at runtime and never appears as a literal here.
	EnvMultiReplica           = "METALDOCS_MULTI_REPLICA"
)

// Store backend values for StoreConfig.Backend.
const (
	StoreBackendMemory = "memory"
	StoreBackendRedis  = "redis"
)

// StoreConfig is the resolved store-backend selection, independent of how it
// was obtained (env vars in production, literal struct values in tests).
// Validate is a pure function over these fields — no env access — so the
// multi-replica guard is directly unit-testable (see TestStoreConfig_MultiReplicaGuard).
type StoreConfig struct {
	// Backend selects the Store implementation: "memory" (default) or
	// "redis". Empty string is treated as "memory".
	Backend string

	// RedisAddr is required when Backend == "redis" (e.g. "127.0.0.1:6379").
	RedisAddr string

	// RedisPassword is optional Redis AUTH.
	RedisPassword string

	// MultiReplica declares operator intent that metaldocs-api runs with
	// more than one replica. When true, Backend must resolve to "redis" —
	// the in-memory backend is per-process and silently multiplies the
	// effective quota by the replica count (the defect this task fixes).
	MultiReplica bool
}

// resolvedBackend returns Backend, defaulting empty to StoreBackendMemory.
func (c StoreConfig) resolvedBackend() string {
	if strings.TrimSpace(c.Backend) == "" {
		return StoreBackendMemory
	}
	return c.Backend
}

// Validate enforces the startup guard: multi-replica intent combined with
// the in-memory store (explicit or default) must refuse to boot, since each
// replica's in-memory counters are independent (N-replica quota
// multiplication). Also validates the redis backend has an address, and
// rejects unknown backend names. Pure function — no env reads — so it is
// directly unit-testable without constructing a real Store or network
// connection.
func (c StoreConfig) Validate() error {
	backend := c.resolvedBackend()
	switch backend {
	case StoreBackendMemory:
		if c.MultiReplica {
			return fmt.Errorf(
				"ratelimit: %s=true requires %s=redis (in-memory store is per-process; "+
					"each replica would enforce an independent quota, silently multiplying "+
					"the effective fleet-wide limit by the replica count)",
				EnvMultiReplica, EnvRatelimitStore,
			)
		}
		return nil
	case StoreBackendRedis:
		if strings.TrimSpace(c.RedisAddr) == "" {
			return fmt.Errorf("ratelimit: %s=redis requires %s to be set", EnvRatelimitStore, EnvRatelimitRedisAddr)
		}
		return nil
	default:
		return fmt.Errorf("ratelimit: unknown %s value %q (want %q or %q)",
			EnvRatelimitStore, c.Backend, StoreBackendMemory, StoreBackendRedis)
	}
}

// LoadStoreConfig reads the store-backend env vars via getenv (injected so
// callers/tests need not touch real process env — see CLAUDE.md's "never
// read/print .env" rule; this reads regular process env vars, not .env
// files, via whatever getenv the caller supplies, typically os.Getenv) and
// returns a validated StoreConfig. Returns the same error Validate would
// return for an invalid combination — main.go should log-and-exit on error,
// matching the existing loginRateCfg error-handling pattern.
func LoadStoreConfig(getenv func(string) string) (StoreConfig, error) {
	cfg := StoreConfig{
		Backend:       strings.TrimSpace(getenv(EnvRatelimitStore)),
		RedisAddr:     strings.TrimSpace(getenv(EnvRatelimitRedisAddr)),
		RedisPassword: getenv(EnvRatelimitRedisPassword),
		MultiReplica:  parseBool(getenv(EnvMultiReplica)),
	}
	if err := cfg.Validate(); err != nil {
		return StoreConfig{}, err
	}
	// Normalize so callers can switch on Backend directly (empty → "memory").
	cfg.Backend = cfg.resolvedBackend()
	return cfg, nil
}

func parseBool(s string) bool {
	b, err := strconv.ParseBool(strings.TrimSpace(s))
	if err != nil {
		return false
	}
	return b
}
