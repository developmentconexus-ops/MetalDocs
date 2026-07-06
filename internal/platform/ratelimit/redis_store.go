package ratelimit

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RedisStoreOptions configures the Redis-backed Store. Zero value is valid
// (nil logger falls back to slog.Default()).
type RedisStoreOptions struct {
	Logger *slog.Logger
}

// redisStore implements Store on top of go-redis + redis_rate's GCRA
// (Generic Cell Rate Algorithm) limiter. Because the counting state lives in
// Redis rather than process memory, every replica sharing the same Redis
// instance/cluster enforces ONE combined quota — this is the fix for the
// N-replica quota-multiplication defect that memoryStore cannot avoid.
//
// GCRA is burst-tolerant by design: redis_rate.PerMinute(quota) sets
// Burst == quota, meaning a cold key can admit up to `quota` requests
// instantly before the per-request spacing kicks in. That is expected GCRA
// behavior (not a bug) and is exercised/documented in
// TestCrossReplicaSharedBudget's redis_shared_budget subtest.
type redisStore struct {
	client *redis.Client
	rate   *redis_rate.Limiter
	logger *slog.Logger

	// ownsClient marks that this store created its client (via
	// NewStoreFromConfig) and must close it on Close. Stores built by
	// NewRedisStore wrap a caller-owned client and never close it.
	ownsClient bool
}

// NewRedisStore constructs a Store backed by client using GCRA rate limiting
// (redis_rate). client is not closed by Store.Close (the caller may share
// one client across multiple Middleware/Store instances — see
// StoreConfig-driven main.go wiring — and owns its lifecycle).
func NewRedisStore(client *redis.Client, opts *RedisStoreOptions) Store {
	logger := slog.Default()
	if opts != nil && opts.Logger != nil {
		logger = opts.Logger
	}
	return &redisStore{
		client: client,
		rate:   redis_rate.NewLimiter(client),
		logger: logger,
	}
}

// Allow implements Store via GCRA. Fail-open on Redis error: see the comment
// on the error branch below for the rationale.
func (s *redisStore) Allow(ctx context.Context, key string, quota int, window time.Duration) (bool, time.Duration, error) {
	limit := toGCRALimit(quota, window)
	res, err := s.rate.Allow(ctx, key, limit)
	if err != nil {
		// Fail-open on Redis error (network error, timeout, connection
		// refused, etc.): a rate limiter's job is to protect the system from
		// overload, not to be a security boundary. When the limiter's own
		// backing store is unavailable, availability wins over strictness —
		// denying every request because Redis hiccuped would turn a
		// dependency blip into a full outage, which is a worse failure mode
		// than temporarily running unlimited. This trade-off is intentional
		// and reviewed: MetalDocs M8 F8.2 Task A, ADR 0071 (wiki/decisions/
		// 0071-distributed-rate-limiting-shared-store.md) §Decision item 5.
		s.logger.Warn("ratelimit: redis error, failing open", "err", err, "key", key)
		return true, 0, nil
	}
	if res.Allowed > 0 {
		return true, 0, nil
	}
	retryAfter := res.RetryAfter
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	return false, retryAfter, nil
}

// Close releases the Redis connection when this store owns it (constructed
// via NewStoreFromConfig). No-op for stores wrapping a caller-owned client
// (NewRedisStore).
func (s *redisStore) Close() {
	if s.ownsClient && s.client != nil {
		_ = s.client.Close()
	}
}

// NewStoreFromConfig builds the process-wide shared Store selected by sc.
//
// Memory backend: returns (nil, nil) — callers leave Config.Store nil so each
// Middleware constructs its own private in-memory store, preserving the
// pre-existing single-replica behavior exactly.
//
// Redis backend: dials ONE Redis client (owned by the returned Store; Close
// releases it) that main.go shares across both limiter mounts (pre-auth login
// + global envelope), so the whole process holds exactly one Redis
// connection pool. Redis stays an implementation detail of this package —
// callers never import go-redis (task constraint: no Redis use outside
// ratelimit).
func NewStoreFromConfig(sc StoreConfig, logger *slog.Logger) (Store, error) {
	if err := sc.Validate(); err != nil {
		return nil, err
	}
	if sc.resolvedBackend() != StoreBackendRedis {
		return nil, nil
	}
	client := redis.NewClient(&redis.Options{
		Addr:     sc.RedisAddr,
		Password: sc.RedisPassword,
	})
	s := NewRedisStore(client, &RedisStoreOptions{Logger: logger}).(*redisStore)
	s.ownsClient = true
	return s, nil
}

// toGCRALimit maps a per-minute quota to a redis_rate.Limit. quota requests
// are allowed per window (today always time.Minute); Burst == quota, matching
// redis_rate.PerMinute's own definition for a 1-minute window. Kept as an
// explicit helper (rather than calling redis_rate.PerMinute directly) so a
// non-minute window is handled correctly, though Store's current only caller
// (Middleware.Limit) always passes time.Minute.
func toGCRALimit(quota int, window time.Duration) redis_rate.Limit {
	if window == time.Minute {
		return redis_rate.PerMinute(quota)
	}
	return redis_rate.Limit{Rate: quota, Period: window, Burst: quota}
}
