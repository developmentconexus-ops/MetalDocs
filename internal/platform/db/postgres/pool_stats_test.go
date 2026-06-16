package postgres_test

import (
	"database/sql"
	"testing"

	"metaldocs/internal/platform/db/postgres"
)

func TestSQLDBPoolStatsAdapter_Maps(t *testing.T) {
	stats := sql.DBStats{
		MaxOpenConnections: 10,
		OpenConnections:    4,
		InUse:              2,
		Idle:               2,
		WaitCount:          5,
	}
	adapter := postgres.NewPoolStatsAdapterFromStats(stats)
	m := adapter.DBPoolStats()

	cases := []struct {
		key  string
		want int
	}{
		{"max_open_connections", 10},
		{"open_connections", 4},
		{"in_use", 2},
		{"idle", 2},
	}
	for _, tc := range cases {
		v, ok := m[tc.key]
		if !ok {
			t.Errorf("missing key %q", tc.key)
			continue
		}
		// Values may be int or int64 depending on field type
		switch got := v.(type) {
		case int:
			if got != tc.want {
				t.Errorf("key %q: want %d got %d", tc.key, tc.want, got)
			}
		case int64:
			if int(got) != tc.want {
				t.Errorf("key %q: want %d got %d", tc.key, tc.want, got)
			}
		default:
			t.Errorf("key %q: unexpected type %T value %v", tc.key, v, v)
		}
	}

	if _, ok := m["wait_count"]; !ok {
		t.Error("missing key wait_count")
	}
}
