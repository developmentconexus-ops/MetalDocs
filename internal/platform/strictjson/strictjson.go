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
	ErrContentType  = errors.New("content-type must be application/json")
	ErrBodyTooLarge = errors.New("request body too large (max 64 KB)")
	ErrEmptyBody    = errors.New("request body must not be empty")
)

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
