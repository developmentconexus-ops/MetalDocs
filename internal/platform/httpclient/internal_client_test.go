package httpclient_test

import (
	"net/http"
	"testing"
	"time"

	"metaldocs/internal/platform/httpclient"
)

func TestNewInternalClient(t *testing.T) {
	c := httpclient.NewInternalClient()

	if c == nil {
		t.Fatal("NewInternalClient returned nil")
	}
	if c.Timeout != 60*time.Second {
		t.Errorf("Client.Timeout = %v, want 60s", c.Timeout)
	}

	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport type = %T, want *http.Transport", c.Transport)
	}

	if tr.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want 5s", tr.TLSHandshakeTimeout)
	}
	if tr.ResponseHeaderTimeout != 10*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 10s", tr.ResponseHeaderTimeout)
	}
	if tr.ExpectContinueTimeout != 1*time.Second {
		t.Errorf("ExpectContinueTimeout = %v, want 1s", tr.ExpectContinueTimeout)
	}
	if tr.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want 90s", tr.IdleConnTimeout)
	}
	if tr.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", tr.MaxIdleConns)
	}
	if tr.MaxIdleConnsPerHost != 20 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 20", tr.MaxIdleConnsPerHost)
	}
	if tr.MaxConnsPerHost != 50 {
		t.Errorf("MaxConnsPerHost = %d, want 50", tr.MaxConnsPerHost)
	}
	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true")
	}
	if tr.DialContext == nil {
		t.Error("DialContext is nil, want non-nil (5s dial timeout)")
	}
}
