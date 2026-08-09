package ratelimit_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"metaldocs/internal/platform/ratelimit"
)

// redisAddrForTest resolves a Redis backend for tests. Preference order:
//  1. METALDOCS_TEST_REDIS_ADDR (explicit operator-provided Redis)
//  2. 127.0.0.1:6379 (local compose redis)
//  3. miniredis in-process fixture — LABELED fixture-based in the test log.
//
// Both real-Redis and fixture paths run the exact same assertions.
func redisAddrForTest(t *testing.T) string {
	t.Helper()
	addr := os.Getenv("METALDOCS_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	probe := redis.NewClient(&redis.Options{Addr: addr, DialTimeout: 500 * time.Millisecond})
	defer func() { _ = probe.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := probe.Ping(ctx).Err(); err == nil {
		t.Logf("redis backend: real Redis at %s", addr)
		return addr
	}
	mr := miniredis.RunT(t)
	t.Logf("redis backend: FIXTURE-BASED (miniredis at %s) — real Redis unreachable", mr.Addr())
	return mr.Addr()
}

// TestCrossReplicaSharedBudget proves the distributed-store contract: two
// independent Middleware instances (simulating two api replicas) sharing ONE
// Redis backend enforce ONE combined budget, while two in-memory instances
// silently double it (the defect this task fixes).
func TestCrossReplicaSharedBudget(t *testing.T) {
	const quota = 10 // req/min
	const requests = 30

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })

	t.Run("redis_shared_budget", func(t *testing.T) {
		addr := redisAddrForTest(t)
		// Two separate clients to the same Redis — one per "replica".
		c1 := redis.NewClient(&redis.Options{Addr: addr})
		c2 := redis.NewClient(&redis.Options{Addr: addr})
		t.Cleanup(func() { _ = c1.Close(); _ = c2.Close() })

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		cfg1, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: quota})
		if err != nil {
			t.Fatalf("NewConfig: %v", err)
		}
		cfg1.Store = ratelimit.NewRedisStore(c1, nil)
		cfg2, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: quota})
		if err != nil {
			t.Fatalf("NewConfig: %v", err)
		}
		cfg2.Store = ratelimit.NewRedisStore(c2, nil)

		mw1 := ratelimit.New(ctx, cfg1)
		mw2 := ratelimit.New(ctx, cfg2)

		// Unique identity per run so state on a real (shared) Redis from a
		// prior run cannot bleed in.
		user := fmt.Sprintf("cross-replica-%d", time.Now().UnixNano())
		extractor := func(r *http.Request) string { return user }
		h1 := mw1.Limit(ratelimit.RouteExportPDF, extractor, next)
		h2 := mw2.Limit(ratelimit.RouteExportPDF, extractor, next)

		admitted := 0
		for i := 0; i < requests; i++ {
			h := h1
			if i%2 == 1 {
				h = h2
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest("POST", "/x", nil))
			if rr.Code == 204 {
				admitted++
			} else if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("req %d: want 204 or 429, got %d", i, rr.Code)
			}
		}

		// GCRA burst tolerance: redis_rate uses GCRA with burst == quota, so a
		// cold key admits exactly `quota` requests instantly. Tokens regenerate
		// at quota/minute (1 per 6s at quota=10); this loop completes in
		// milliseconds, so at most 1 extra token can regenerate mid-loop on a
		// slow machine. Documented tolerance: +1.
		const gcraBurstTolerance = 1
		if admitted > quota+gcraBurstTolerance {
			t.Fatalf("combined admits across 2 replicas sharing one Redis: want <= %d (quota %d + GCRA tolerance %d), got %d",
				quota+gcraBurstTolerance, quota, gcraBurstTolerance, admitted)
		}
		if admitted < quota {
			t.Fatalf("combined admits: want >= quota %d, got %d (limiter over-strict)", quota, admitted)
		}
		t.Logf("redis shared budget: %d/%d admitted (quota %d, tolerance +%d)", admitted, requests, quota, gcraBurstTolerance)
	})

	// Contrast subtest: pins the defect this task fixes — two in-memory
	// replicas each hold an independent budget, so the combined admit count is
	// 2×quota, silently doubling the intended limit.
	t.Run("memory_pins_defect_double_budget", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		mkHandler := func() http.Handler {
			cfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteExportPDF: quota})
			if err != nil {
				t.Fatalf("NewConfig: %v", err)
			}
			mw := ratelimit.New(ctx, cfg) // default in-memory store
			return mw.Limit(ratelimit.RouteExportPDF, func(r *http.Request) string { return "same-user" }, next)
		}
		h1 := mkHandler()
		h2 := mkHandler()

		admitted := 0
		for i := 0; i < requests; i++ {
			h := h1
			if i%2 == 1 {
				h = h2
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest("POST", "/x", nil))
			if rr.Code == 204 {
				admitted++
			}
		}
		if admitted != 2*quota {
			t.Fatalf("two in-memory replicas: want exactly %d combined admits (2x quota — the defect), got %d", 2*quota, admitted)
		}
		t.Logf("memory contrast: %d/%d admitted — double the intended quota of %d (defect pinned)", admitted, requests, quota)
	})
}

// TestStoreConfig_MultiReplicaGuard proves the startup guard: multi-replica
// intent combined with the (per-process) memory store must refuse to boot.
func TestStoreConfig_MultiReplicaGuard(t *testing.T) {
	cases := []struct {
		name    string
		cfg     ratelimit.StoreConfig
		wantErr bool
	}{
		{"multi-replica + explicit memory → error",
			ratelimit.StoreConfig{Backend: "memory", MultiReplica: true}, true},
		{"multi-replica + default (empty) backend → error",
			ratelimit.StoreConfig{Backend: "", MultiReplica: true}, true},
		{"multi-replica + redis with addr → ok",
			ratelimit.StoreConfig{Backend: "redis", RedisAddr: "127.0.0.1:6379", MultiReplica: true}, false},
		{"single replica + memory → ok",
			ratelimit.StoreConfig{Backend: "memory", MultiReplica: false}, false},
		{"single replica + default backend → ok",
			ratelimit.StoreConfig{Backend: "", MultiReplica: false}, false},
		{"redis without addr → error",
			ratelimit.StoreConfig{Backend: "redis", RedisAddr: "", MultiReplica: false}, true},
		{"unknown backend → error",
			ratelimit.StoreConfig{Backend: "bogus", MultiReplica: false}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("want error, got nil for %+v", tc.cfg)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("want nil error, got %v for %+v", tc.cfg, err)
			}
		})
	}
}

// TestLoadStoreConfig proves env parsing: variable names are the shipped
// contract (METALDOCS_RATELIMIT_STORE, METALDOCS_RATELIMIT_REDIS_ADDR,
// METALDOCS_RATELIMIT_REDIS_PASSWORD, METALDOCS_MULTI_REPLICA).
func TestLoadStoreConfig(t *testing.T) {
	env := map[string]string{
		"METALDOCS_RATELIMIT_STORE":          "redis",
		"METALDOCS_RATELIMIT_REDIS_ADDR":     "10.0.0.5:6379",
		"METALDOCS_RATELIMIT_REDIS_PASSWORD": "s3cret",
		"METALDOCS_MULTI_REPLICA":            "true",
	}
	cfg, err := ratelimit.LoadStoreConfig(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("LoadStoreConfig: %v", err)
	}
	if cfg.Backend != "redis" || cfg.RedisAddr != "10.0.0.5:6379" || cfg.RedisPassword != "s3cret" || !cfg.MultiReplica {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	// Defaults: everything unset → memory backend, single replica, valid.
	cfg, err = ratelimit.LoadStoreConfig(func(string) string { return "" })
	if err != nil {
		t.Fatalf("LoadStoreConfig defaults: %v", err)
	}
	if cfg.Backend != "memory" || cfg.MultiReplica {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	// The guard fires through LoadStoreConfig too (startup path).
	_, err = ratelimit.LoadStoreConfig(func(k string) string {
		if k == "METALDOCS_MULTI_REPLICA" {
			return "true"
		}
		return ""
	})
	if err == nil {
		t.Fatal("multi_replica=true with default memory store: want boot-refusing error, got nil")
	}
}
