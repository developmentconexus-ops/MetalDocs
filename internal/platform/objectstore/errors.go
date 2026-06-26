package objectstore

import "errors"

// Kernel-owned sentinels. No module-domain imports; each module maps these to its
// own domain error at the application boundary.
var (
	ErrObjectMissing    = errors.New("objectstore: object not found")
	ErrHashMismatch     = errors.New("objectstore: content hash mismatch")
	ErrObjectTooLarge   = errors.New("objectstore: object exceeds max size")
	ErrKeyOutsideTenant = errors.New("objectstore: key outside tenant scope")
)
