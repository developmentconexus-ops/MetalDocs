package objectstore

import (
	"errors"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
)

// Kernel-owned sentinels. No module-domain imports; each module maps these to its
// own domain error at the application boundary.
var (
	ErrObjectMissing    = errors.New("objectstore: object not found")
	ErrHashMismatch     = errors.New("objectstore: content hash mismatch")
	ErrObjectTooLarge   = errors.New("objectstore: object exceeds max size")
	ErrKeyOutsideTenant = errors.New("objectstore: key outside tenant scope")
)

// isNoSuchKeyErr reports whether err represents a missing-object (NoSuchKey) response
// from MinIO / S3. It is package-private; callers within this package translate it to
// ErrObjectMissing at their own boundary.
func isNoSuchKeyErr(err error) bool {
	if err == nil {
		return false
	}
	var resp minio.ErrorResponse
	if errors.As(err, &resp) && strings.EqualFold(resp.Code, "NoSuchKey") {
		return true
	}
	if strings.Contains(err.Error(), "NoSuchKey") {
		return true
	}
	var ue *url.Error
	if errors.As(err, &ue) && strings.Contains(ue.Error(), "NoSuchKey") {
		return true
	}
	return false
}
