package postgres

import (
	"database/sql"
	"time"
)

// SQLDBPoolStatsAdapter wraps *sql.DB and implements observability.DBPoolStatsProvider.
type SQLDBPoolStatsAdapter struct {
	db *sql.DB
}

// NewPoolStatsAdapter wraps db for use as a DBPoolStatsProvider.
// Returns nil when db is nil.
func NewPoolStatsAdapter(db *sql.DB) *SQLDBPoolStatsAdapter {
	if db == nil {
		return nil
	}
	return &SQLDBPoolStatsAdapter{db: db}
}

// DBPoolStats returns live pool stats as a map[string]any with snake_case keys.
func (a *SQLDBPoolStatsAdapter) DBPoolStats() map[string]any {
	return dbStatsToMap(a.db.Stats())
}

// sqlDBStatsAdapter is the test-only variant backed by a pre-built sql.DBStats.
type sqlDBStatsAdapter struct {
	stats sql.DBStats
}

// NewPoolStatsAdapterFromStats builds an adapter from a known sql.DBStats value.
// Used in tests where a real *sql.DB is not available.
func NewPoolStatsAdapterFromStats(stats sql.DBStats) *sqlDBStatsAdapter {
	return &sqlDBStatsAdapter{stats: stats}
}

func (a *sqlDBStatsAdapter) DBPoolStats() map[string]any {
	return dbStatsToMap(a.stats)
}

func dbStatsToMap(s sql.DBStats) map[string]any {
	return map[string]any{
		"max_open_connections": s.MaxOpenConnections,
		"open_connections":     s.OpenConnections,
		"in_use":               s.InUse,
		"idle":                 s.Idle,
		"wait_count":           s.WaitCount,
		"wait_duration_ms":     s.WaitDuration / time.Millisecond,
		"max_idle_closed":      s.MaxIdleClosed,
		"max_idle_time_closed": s.MaxIdleTimeClosed,
		"max_lifetime_closed":  s.MaxLifetimeClosed,
	}
}
