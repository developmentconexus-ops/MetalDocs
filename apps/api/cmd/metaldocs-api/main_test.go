package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/objectstore"
)

// TestShutdownServer_ServerErrorReturnsExitOne reproduces the C3 regression:
// when ListenAndServe returns a non-ErrServerClosed error, the teardown path
// must report a non-zero exit code instead of os.Exit-ing straight through.
// shutdownServer no longer joins a worker WaitGroup — the staging outbox
// poll workers it used to join were retired (M5 F5.3 T4; River now owns
// dispatch retry) and the scheduler WaitGroup was retired earlier still
// (M5 F5.2 T4, ADR 0067).
func TestShutdownServer_ServerErrorReturnsExitOne(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	server := &http.Server{Addr: "127.0.0.1:0"}
	metricsServer := &http.Server{Addr: "127.0.0.1:0"}
	serverErr := make(chan error, 1)
	serverErr <- errors.New("simulated listen failure")
	metricsErr := make(chan error, 1) // intentionally empty

	done := make(chan int, 1)
	go func() {
		done <- shutdownServer(ctx, stop, server, metricsServer, serverErr, metricsErr)
	}()

	select {
	case exitCode := <-done:
		if exitCode != 1 {
			t.Fatalf("expected exit code 1 on server error, got %d", exitCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdownServer did not return within 5s")
	}
}

// TestShutdownServer_ErrServerClosedReturnsZero confirms the canonical
// "server.Shutdown caused ListenAndServe to return" path is not treated as
// a failure.
func TestShutdownServer_ErrServerClosedReturnsZero(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	server := &http.Server{Addr: "127.0.0.1:0"}
	metricsServer := &http.Server{Addr: "127.0.0.1:0"}
	serverErr := make(chan error, 1)
	serverErr <- http.ErrServerClosed
	metricsErr := make(chan error, 1) // intentionally empty

	exitCode := shutdownServer(ctx, stop, server, metricsServer, serverErr, metricsErr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for ErrServerClosed, got %d", exitCode)
	}
}

// TestShutdownServer_ContextCancellation confirms the Ctrl-C path returns
// exit code 0.
func TestShutdownServer_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	server := &http.Server{Addr: "127.0.0.1:0"}
	metricsServer := &http.Server{Addr: "127.0.0.1:0"}
	serverErr := make(chan error, 1)  // intentionally empty
	metricsErr := make(chan error, 1) // intentionally empty

	stop() // simulate SIGTERM arriving before any listen error

	exitCode := shutdownServer(ctx, stop, server, metricsServer, serverErr, metricsErr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0 on context cancel, got %d", exitCode)
	}
}

func TestBuildTemplatesModuleFailsWhenCapabilityServiceMissing(t *testing.T) {
	t.Parallel()

	presigner := objectstore.NewVerifiedStore(nil, nil, "metaldocs-attachments", 0)
	_, _, _, err := buildTemplatesModule(bootstrap.APIDependencies{}, nil, presigner, iamdomain.NoopUserDisplayNameReader{})
	if err == nil {
		t.Fatal("expected error for nil capability service")
	}
}

// TestBuildTemplatesModuleFailsWhenPresignerMissing guards the STO-02 wiring
// change: buildTemplatesModule no longer constructs its own VerifiedStore, so a
// nil presigner must fail fast at construction instead of surfacing as a later
// nil-pointer panic on first object-store use.
func TestBuildTemplatesModuleFailsWhenPresignerMissing(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildTemplatesModule(bootstrap.APIDependencies{}, nil, nil, iamdomain.NoopUserDisplayNameReader{})
	if err == nil {
		t.Fatal("expected error for nil presigner")
	}
}

func TestRequirePostgresRepositoryModeRejectsMemory(t *testing.T) {
	t.Parallel()

	err := requirePostgresRepositoryMode("memory")
	if err == nil {
		t.Fatal("expected memory mode to be rejected")
	}
}

func TestRequirePostgresRepositoryModeAllowsPostgres(t *testing.T) {
	t.Parallel()

	if err := requirePostgresRepositoryMode("postgres"); err != nil {
		t.Fatalf("postgres mode should be allowed: %v", err)
	}
}

func TestRequireApprovalRuntimeSupport_RejectsMissingFanoutURL(t *testing.T) {
	t.Parallel()

	if err := requireApprovalRuntimeSupport(""); err == nil {
		t.Fatal("expected missing fanout URL to be rejected")
	}
}

func TestRequireApprovalRuntimeSupport_AllowsConfiguredFanoutURL(t *testing.T) {
	t.Parallel()

	if err := requireApprovalRuntimeSupport("http://fanout.internal"); err != nil {
		t.Fatalf("configured fanout URL should be allowed: %v", err)
	}
}
