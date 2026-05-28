package servicebus

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDocgenV2ClientValidateTemplateReturnsReadError(t *testing.T) {
	client := NewDocgenV2Client("http://docgen.internal", "tok", 0)
	client.http = &http.Client{Transport: validateRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       &validateErrorReadCloser{err: errors.New("boom")},
			Header:     make(http.Header),
		}, nil
	})}

	ok, raw, err := client.ValidateTemplate(context.Background(), "docx", "schema")
	if err == nil {
		t.Fatal("expected read error")
	}
	if ok {
		t.Fatal("expected validation failure on read error")
	}
	if raw != nil {
		t.Fatalf("raw = %q, want nil", raw)
	}
	if !strings.Contains(err.Error(), "read body") {
		t.Fatalf("expected read body context, got %v", err)
	}
}

func TestDocgenV2ClientValidateTemplateRejectsOversizedResponse(t *testing.T) {
	client := NewDocgenV2Client("http://docgen.internal", "tok", 0)
	client.http = &http.Client{Transport: validateRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       io.NopCloser(newValidateRepeatingReader(maxValidateResponseBodyBytes + 1)),
			Header:     make(http.Header),
		}, nil
	})}

	ok, _, err := client.ValidateTemplate(context.Background(), "docx", "schema")
	if err == nil {
		t.Fatal("expected oversized response error")
	}
	if ok {
		t.Fatal("expected validation to fail when body exceeds limit")
	}
	if !strings.Contains(err.Error(), "body exceeds") {
		t.Fatalf("expected size limit error, got %v", err)
	}
}

type validateRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn validateRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type validateErrorReadCloser struct {
	err error
}

func (r *validateErrorReadCloser) Read([]byte) (int, error) {
	return 0, r.err
}

func (r *validateErrorReadCloser) Close() error {
	return nil
}

type validateRepeatingReader struct {
	remaining int64
}

func newValidateRepeatingReader(size int64) *validateRepeatingReader {
	return &validateRepeatingReader{remaining: size}
}

func (r *validateRepeatingReader) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	for i := range p {
		p[i] = 'a'
	}
	r.remaining -= int64(len(p))
	return len(p), nil
}
