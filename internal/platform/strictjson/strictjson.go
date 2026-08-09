// Package strictjson provides RFC-strict request body decoding shared across
// modules: application/json only, bounded size, single JSON value, and
// DisallowUnknownFields. Promoted from the documents approval module so there
// is one source (reuse-don't-reinvent; cross-module import rule).
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
)

const maxRequestBodyBytes = 64 * 1024

var (
	// ErrContentType is returned when the request's Content-Type is not application/json.
	ErrContentType = errors.New("content-type must be application/json")
	// ErrBodyTooLarge is returned when the request body exceeds the 64 KB cap.
	ErrBodyTooLarge = errors.New("request body too large (max 64 KB)")
	// ErrEmptyBody is returned when the request body is missing or empty.
	ErrEmptyBody = errors.New("request body must not be empty")
)

// Decode strictly decodes r's JSON body into dst: requires
// application/json, enforces a 64 KB size cap, rejects unknown fields, and
// rejects a body containing more than one JSON value.
func Decode(r *http.Request, dst any) error {
	if r == nil || r.Body == nil {
		return ErrEmptyBody
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return ErrContentType
	}
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes+1))
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	if len(payload) > maxRequestBodyBytes {
		return ErrBodyTooLarge
	}
	if len(bytes.TrimSpace(payload)) == 0 {
		return ErrEmptyBody
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}
		return err
	}
	return nil
}
