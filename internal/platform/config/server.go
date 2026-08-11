package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// ServerConfig holds HTTP listen configuration.
//
// Two listeners, deliberately separate (F-R1, Dim-9 isolation):
//   - Addr        — the public API listener (APP_PORT).
//   - MetricsAddr — a dedicated Prometheus scrape listener (METRICS_ADDR). The
//     public listener never serves /metrics, so the scrape surface is isolated
//     by process topology, not by ops/ingress discipline. The metrics port is
//     meant to be reachable only on the infra network (not host-published).
type ServerConfig struct {
	Addr        string
	MetricsAddr string
}

// LoadServerConfig reads APP_PORT and METRICS_ADDR from the environment.
//   - APP_PORT unset → Addr ":8080"; if set, must be a port in 1..65535.
//   - METRICS_ADDR unset → MetricsAddr ":9090"; if set, must be a valid listen
//     address (":port" or "host:port", port in 1..65535).
func LoadServerConfig() (ServerConfig, error) {
	cfg := ServerConfig{Addr: ":8080", MetricsAddr: ":9090"}

	if appPort := strings.TrimSpace(os.Getenv("APP_PORT")); appPort != "" {
		port, err := strconv.Atoi(appPort)
		if err != nil || port < 1 || port > 65535 {
			return ServerConfig{}, fmt.Errorf("invalid APP_PORT value")
		}
		cfg.Addr = ":" + strconv.Itoa(port)
	}

	if metricsAddr := strings.TrimSpace(os.Getenv("METRICS_ADDR")); metricsAddr != "" {
		if err := validateListenAddr(metricsAddr); err != nil {
			return ServerConfig{}, fmt.Errorf("invalid METRICS_ADDR value: %w", err)
		}
		cfg.MetricsAddr = metricsAddr
	}

	return cfg, nil
}

// LoadListenAddr reads a generic "host:port" listen address from envVar,
// defaulting to defaultAddr when unset. Shared by any binary that needs its
// own dedicated infra-port listener beyond the api binary's own
// APP_PORT/METRICS_ADDR pair (A7.1: metaldocs-worker's WORKER_METRICS_ADDR,
// metaldocs-jobs' JOBS_METRICS_ADDR) — same validateListenAddr rule
// LoadServerConfig already applies to METRICS_ADDR, factored out here so a
// second and third binary do not each re-implement the parsing/validation.
func LoadListenAddr(envVar, defaultAddr string) (string, error) {
	addr := strings.TrimSpace(os.Getenv(envVar))
	if addr == "" {
		return defaultAddr, nil
	}
	if err := validateListenAddr(addr); err != nil {
		return "", fmt.Errorf("invalid %s value: %w", envVar, err)
	}
	return addr, nil
}

// validateListenAddr accepts a Go listen address of the form ":port" or
// "host:port" with the port in 1..65535.
func validateListenAddr(addr string) error {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("not a host:port address")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("port out of range 1..65535")
	}
	return nil
}
