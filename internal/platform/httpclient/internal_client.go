package httpclient

import (
	"net"
	"net/http"
	"time"
)

// NewInternalClient returns an *http.Client tuned for internal service fanout calls.
// Transport settings are sized for low-latency intra-cluster HTTP/2 traffic.
// No retry logic is embedded here; retry is owned by PDFOutboxWorker (wiki/decisions/0009-pdf-dispatch-outbox.md).
func NewInternalClient() *http.Client {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       50,
		ForceAttemptHTTP2:     true,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   60 * time.Second,
	}
}
