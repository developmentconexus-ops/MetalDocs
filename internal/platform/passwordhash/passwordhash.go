// Package passwordhash is a pure, DB-free platform primitive for password
// hashing and verification (REQ-AUTHN-1: passwords hashed with a
// memory-hard KDF, Argon2id family). It is the single home for:
//
//   - Argon2id hashing/verification, PHC string encoded (params travel with
//     the hash, so a future parameter change never orphans existing hashes).
//   - Algorithm dispatch (Verify), so callers holding a stored password_algo
//     discriminator ("bcrypt" or "argon2id") can compare against either
//     without duplicating the switch in every module that checks a
//     password (auth login, approval signature reauth).
//
// Verify fails closed for any algorithm this package does not recognize —
// it never falls through to a different comparison. This package has no
// knowledge of tenants, Postgres, sessions, or any MetalDocs domain
// concept; callers own persistence of the hash string and the
// password_algo column.
package passwordhash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Algorithm discriminators. These are the only values a password_algo
// column may legitimately hold; any other stored value is unrecognized and
// Verify fails closed on it.
const (
	AlgoBcrypt   = "bcrypt"
	AlgoArgon2id = "argon2id"
)

// Argon2id parameters for interactive login.
//
// Memory/parallelism source: OWASP Password Storage Cheat Sheet
// (https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html),
// Argon2id section, the cheat sheet's first-listed configuration for
// interactive login: m=19456 (19 MiB), p=1 (parallelism). This is also
// consistent with RFC 9106's second recommended option family
// (memory-constrained environments).
//
// Time cost is raised to t=16, above the cheat sheet's baseline minimum of
// t=2. Rationale: this is a memory-hard-KDF migration, not a fresh system —
// the pre-existing login path (bcrypt, cost 12) already spent roughly
// 200-300ms of CPU per attempt on this class of hardware, and a migration
// must not silently cut the brute-force cost budget an attacker faces
// (no security regression during migration). At t=2 a single Argon2id call
// measured 25-45ms here, well under the legacy bcrypt(12) cost; t=16 brings
// it to a comparable ~150-300ms while keeping m/p at the cited minimum, so
// the login endpoint's per-attempt cost does not regress relative to the
// system it replaces. This is strictly additional hardening on top of the
// cited minimum, not a deviation from it.
//
// Every stored hash encodes these params in its own PHC string
// ($argon2id$v=19$m=...,t=...,p=...$salt$hash), so verification always uses
// the params a hash was minted with, not the package's current defaults —
// a future parameter bump never invalidates hashes minted under the old ones.
const (
	argon2idTime    uint32 = 16
	argon2idMemory  uint32 = 19 * 1024 // KiB = 19 MiB
	argon2idThreads uint8  = 1
	argon2idSaltLen        = 16
	argon2idKeyLen  uint32 = 32
)

var (
	// ErrUnrecognizedAlgorithm is returned by Verify when the algorithm
	// discriminator is neither "bcrypt" nor "argon2id". No-fallback
	// principle: an unrecognized algorithm never falls through to a
	// different comparison, it is a hard verification failure.
	ErrUnrecognizedAlgorithm = errors.New("passwordhash: unrecognized password algorithm")
	// ErrMalformedHash is returned when a stored value claims to be an
	// argon2id PHC string but cannot be parsed as one.
	ErrMalformedHash = errors.New("passwordhash: malformed argon2id hash")
)

// argon2idParams is the parameter triple encoded in an argon2id PHC string.
type argon2idParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

// HashArgon2id hashes password with this package's current Argon2id
// parameters and a freshly generated random salt, returning the standard
// PHC string form: $argon2id$v=19$m=...,t=...,p=...$salt$hash.
func HashArgon2id(password []byte) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("passwordhash: generate salt: %w", err)
	}
	hash := argon2.IDKey(password, salt, argon2idTime, argon2idMemory, argon2idThreads, argon2idKeyLen)
	return encodeArgon2id(argon2idParams{memory: argon2idMemory, time: argon2idTime, threads: argon2idThreads}, salt, hash), nil
}

func encodeArgon2id(p argon2idParams, salt, hash []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.time, p.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

// parseArgon2id parses a $argon2id$v=..$m=..,t=..,p=..$salt$hash PHC
// string, returning the params it was minted with plus the decoded salt
// and hash bytes.
func parseArgon2id(encoded string) (argon2idParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// parts[0] is empty (string starts with "$"): ["", "argon2id", "v=19", "m=..,t=..,p=..", salt, hash]
	if len(parts) != 6 || parts[1] != AlgoArgon2id {
		return argon2idParams{}, nil, nil, ErrMalformedHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return argon2idParams{}, nil, nil, fmt.Errorf("%w: version: %v", ErrMalformedHash, err)
	}
	if version != argon2.Version {
		return argon2idParams{}, nil, nil, fmt.Errorf("%w: unsupported version %d", ErrMalformedHash, version)
	}
	var m, t, par int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &par); err != nil {
		return argon2idParams{}, nil, nil, fmt.Errorf("%w: params: %v", ErrMalformedHash, err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argon2idParams{}, nil, nil, fmt.Errorf("%w: salt: %v", ErrMalformedHash, err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argon2idParams{}, nil, nil, fmt.Errorf("%w: hash: %v", ErrMalformedHash, err)
	}
	return argon2idParams{memory: uint32(m), time: uint32(t), threads: uint8(par)}, salt, hash, nil
}

// VerifyArgon2id parses an argon2id PHC-format hash and compares it against
// password, computing the candidate hash with the parameters ENCODED IN
// THE STORED HASH (never this package's current defaults), so a future
// parameter bump never invalidates hashes minted under the old ones.
// Comparison is constant-time.
func VerifyArgon2id(encoded string, password []byte) (bool, error) {
	params, salt, want, err := parseArgon2id(encoded)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey(password, salt, params.time, params.memory, params.threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// Argon2ParamsSummary returns the "m=...,t=...,p=..." parameter string
// encoded in an argon2id PHC hash, for audit/attestation logging (never for
// verification — VerifyArgon2id always uses the parsed params directly).
func Argon2ParamsSummary(encoded string) (string, error) {
	params, _, _, err := parseArgon2id(encoded)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("m=%d,t=%d,p=%d", params.memory, params.time, params.threads), nil
}

// Verify dispatches on algo — the stored password_algo discriminator — and
// reports whether password matches hash. It fails closed for any algorithm
// value other than AlgoBcrypt/AlgoArgon2id: it returns
// (false, ErrUnrecognizedAlgorithm) rather than guessing or trying a
// different comparison (no-fallback principle for integrity-critical
// reads). A malformed argon2id hash likewise returns (false, err) rather
// than panicking or matching.
func Verify(algo string, hash, password []byte) (bool, error) {
	switch algo {
	case AlgoBcrypt:
		return bcrypt.CompareHashAndPassword(hash, password) == nil, nil
	case AlgoArgon2id:
		return VerifyArgon2id(string(hash), password)
	default:
		return false, fmt.Errorf("%w: %q", ErrUnrecognizedAlgorithm, algo)
	}
}
