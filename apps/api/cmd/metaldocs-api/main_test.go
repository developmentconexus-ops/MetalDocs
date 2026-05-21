package main

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestShutdownServer_ServerErrorJoinsScheduler reproduces the C3 regression:
// when ListenAndServe returns a non-ErrServerClosed error, the teardown path
// must still drain the scheduler goroutine instead of os.Exit-ing and leaking
// the WaitGroup.
func TestShutdownServer_ServerErrorJoinsScheduler(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	var schedulerWG, workerWG sync.WaitGroup
	var schedulerJoined atomic.Bool

	schedulerWG.Add(1)
	go func() {
		defer schedulerWG.Done()
		<-ctx.Done()
		schedulerJoined.Store(true)
	}()

	server := &http.Server{Addr: "127.0.0.1:0"}
	serverErr := make(chan error, 1)
	serverErr <- errors.New("simulated listen failure")

	done := make(chan int, 1)
	go func() {
		done <- shutdownServer(ctx, stop, server, serverErr, &schedulerWG, &workerWG)
	}()

	select {
	case exitCode := <-done:
		if exitCode != 1 {
			t.Fatalf("expected exit code 1 on server error, got %d", exitCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdownServer did not return within 5s — scheduler likely leaked")
	}

	if !schedulerJoined.Load() {
		t.Fatal("scheduler goroutine was not joined before shutdownServer returned")
	}
}

// TestShutdownServer_ErrServerClosedReturnsZero confirms the canonical
// "server.Shutdown caused ListenAndServe to return" path is not treated as
// a failure.
func TestShutdownServer_ErrServerClosedReturnsZero(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	var schedulerWG, workerWG sync.WaitGroup
	server := &http.Server{Addr: "127.0.0.1:0"}
	serverErr := make(chan error, 1)
	serverErr <- http.ErrServerClosed

	exitCode := shutdownServer(ctx, stop, server, serverErr, &schedulerWG, &workerWG)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for ErrServerClosed, got %d", exitCode)
	}
}

// TestShutdownServer_ContextCancellation confirms the Ctrl-C path returns
// exit code 0 and drains both worker and scheduler WaitGroups.
func TestShutdownServer_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	var schedulerWG, workerWG sync.WaitGroup
	var workerJoined atomic.Bool

	workerWG.Add(1)
	go func() {
		defer workerWG.Done()
		<-ctx.Done()
		workerJoined.Store(true)
	}()

	server := &http.Server{Addr: "127.0.0.1:0"}
	serverErr := make(chan error, 1) // intentionally empty

	stop() // simulate SIGTERM arriving before any listen error

	exitCode := shutdownServer(ctx, stop, server, serverErr, &schedulerWG, &workerWG)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0 on context cancel, got %d", exitCode)
	}
	if !workerJoined.Load() {
		t.Fatal("worker goroutine was not joined before shutdownServer returned")
	}
}
