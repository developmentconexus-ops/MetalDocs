package strictjson_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/strictjson"
)

type payload struct {
	Name string `json:"name"`
}

func newReq(t *testing.T, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestDecode_OK(t *testing.T) {
	var p payload
	if err := strictjson.Decode(newReq(t, `{"name":"x"}`), &p); err != nil {
		t.Fatalf("Decode = %v, want nil", err)
	}
	if p.Name != "x" {
		t.Fatalf("Name = %q, want x", p.Name)
	}
}

func TestDecode_UnknownFieldRejected(t *testing.T) {
	var p payload
	if err := strictjson.Decode(newReq(t, `{"name":"x","extra":1}`), &p); err == nil {
		t.Fatal("Decode = nil, want error for unknown field")
	}
}

func TestDecode_WrongContentType(t *testing.T) {
	var p payload
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"name":"x"}`)))
	if err := strictjson.Decode(r, &p); !errors.Is(err, strictjson.ErrContentType) {
		t.Fatalf("Decode = %v, want ErrContentType", err)
	}
}

func TestDecode_EmptyBody(t *testing.T) {
	var p payload
	if err := strictjson.Decode(newReq(t, ``), &p); !errors.Is(err, strictjson.ErrEmptyBody) {
		t.Fatalf("Decode = %v, want ErrEmptyBody", err)
	}
}

func TestDecode_BodyTooLarge(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 64*1024+1)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	var p payload
	if err := strictjson.Decode(r, &p); !errors.Is(err, strictjson.ErrBodyTooLarge) {
		t.Fatalf("Decode = %v, want ErrBodyTooLarge", err)
	}
}

func TestDecode_MultipleValues(t *testing.T) {
	var p payload
	if err := strictjson.Decode(newReq(t, `{"name":"x"}{"name":"y"}`), &p); err == nil {
		t.Fatal("Decode = nil, want error for multiple JSON values")
	}
}

func TestDecode_MalformedJSON(t *testing.T) {
	var p payload
	if err := strictjson.Decode(newReq(t, `{bad json`), &p); err == nil {
		t.Fatal("Decode = nil, want error for malformed JSON")
	}
}
