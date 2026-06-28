// Package contracts re-exports the platform strict JSON decoder. The decoder
// was promoted to internal/platform/strictjson (single source). This shim keeps
// existing in-module callers stable; new modules import the platform package.
package contracts

import (
	"net/http"

	"metaldocs/internal/platform/strictjson"
)

var (
	ErrContentType  = strictjson.ErrContentType
	ErrBodyTooLarge = strictjson.ErrBodyTooLarge
	ErrEmptyBody    = strictjson.ErrEmptyBody
)

func Decode(r *http.Request, dst any) error { return strictjson.Decode(r, dst) }
