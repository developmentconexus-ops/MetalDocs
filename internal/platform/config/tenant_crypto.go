package config

// tenant_crypto.go — M7 F7.3 Task B: loads the deployment tenant KEK
// (key-encryption key) from METALDOCS_TENANT_KEK. The KEK is
// base64-encoded in the environment and must decode to exactly 32 bytes
// (AES-256). Unset is a valid, first-class state — the composition root
// wires a nil/no-op TenantCrypto path when configured is false, mirroring
// F7.2's TenantKeyProvisioner no-op seam.
//
// Key material is NEVER logged or included in any returned error string —
// errors here report lengths/shapes only, never the decoded bytes.

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

// TenantKEKSize is the required decoded length (bytes) of METALDOCS_TENANT_KEK.
const TenantKEKSize = 32

// ErrTenantKEKWrongLength is returned when METALDOCS_TENANT_KEK decodes to a
// length other than TenantKEKSize.
var ErrTenantKEKWrongLength = errors.New("METALDOCS_TENANT_KEK must decode to 32 bytes")

// LoadTenantKEK reads and validates METALDOCS_TENANT_KEK. configured=false
// (kek=nil, err=nil) means the env var is unset or blank — a valid state,
// not an error; callers wire the nil/no-op TenantCrypto path in that case.
func LoadTenantKEK() (kek []byte, configured bool, err error) {
	raw := strings.TrimSpace(os.Getenv("METALDOCS_TENANT_KEK"))
	if raw == "" {
		return nil, false, nil
	}
	decoded, decodeErr := base64.StdEncoding.DecodeString(raw)
	if decodeErr != nil {
		return nil, false, fmt.Errorf("invalid METALDOCS_TENANT_KEK: not valid base64: %w", decodeErr)
	}
	if len(decoded) != TenantKEKSize {
		return nil, false, fmt.Errorf("%w (got %d bytes)", ErrTenantKEKWrongLength, len(decoded))
	}
	return decoded, true, nil
}
