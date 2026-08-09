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
	"sync/atomic"

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

// Bounds on the m/t/p triple a stored argon2id PHC string may encode.
//
// These exist to reject garbage, not to reject the system's own hashes or a
// future parameter bump: the current minting params (argon2idMemory=19456,
// argon2idTime=16, argon2idThreads=1, above) sit far inside every bound
// here, with headroom for the params to grow substantially before a bound
// would need revisiting.
//
//   - maxArgon2idMemoryKiB caps the slice argon2.IDKey allocates. A stored
//     hash with an attacker-controlled negative m parses via Sscanf's signed
//     %d with no error, and casting that straight to uint32 (as this parser
//     used to) wraps to near math.MaxUint32 KiB - a ~4 TiB allocation
//     request. That is a Go runtime OOM (runtime.throw), which panic_recovery
//     middleware cannot catch: it kills the whole metaldocs-api process, not
//     just the request. 1 GiB is two orders of magnitude above the 19 MiB
//     minting default - room for real parameter growth - while still being
//     an allocation size any host running this service can satisfy without
//     risking the process.
//   - The lower bound on m is derived as 8*p (not a fixed constant), which
//     argon2.IDKey itself requires (see golang.org/x/crypto/argon2); a hash
//     violating it is malformed regardless of sign.
//   - maxArgon2idTime bounds the CPU a single verify can burn. t is a
//     multiplicative pass count over the memory block; unbounded t lets one
//     stored hash pin a CPU core indefinitely on every login attempt against
//     that account. 64 is 4x the t=16 minting default, generous for a future
//     bump, still boundable CPU time.
//   - maxArgon2idParallelism keeps p inside uint8 (the wire type) and bounds
//     the goroutine fan-out argon2.IDKey spins up per call; p=1 today, 64
//     leaves headroom without letting a single verify spawn hundreds of
//     goroutines.
const (
	maxArgon2idMemoryKiB   = 1 << 20 // 1 GiB
	minArgon2idTime        = 1
	maxArgon2idTime        = 64
	minArgon2idParallelism = 1
	maxArgon2idParallelism = 64
)

// argon2idParams is the parameter triple encoded in an argon2id PHC string.
type argon2idParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

// kdfInvocations counts every real KDF computation this package performs —
// each argon2.IDKey call (HashArgon2id, VerifyArgon2id) and each
// bcrypt.CompareHashAndPassword call (Verify's bcrypt branch). It exists so
// callers with no other way to observe "did the expensive comparison
// actually run" (e.g. a timing-oracle regression test that must not assert
// on wall-clock) can assert on invocation counts instead. It only counts;
// it cannot be used to skip, weaken, or reconfigure the KDF itself — every
// increment sits directly next to the real compute call it measures, not a
// substitute for it.
var kdfInvocations atomic.Uint64

// KDFInvocations returns the number of KDF computations (Argon2id or
// bcrypt) this process has performed since start. Intended for tests that
// need a deterministic substitute for wall-clock timing assertions.
func KDFInvocations() uint64 {
	return kdfInvocations.Load()
}

// HashArgon2id hashes password with this package's current Argon2id
// parameters and a freshly generated random salt, returning the standard
// PHC string form: $argon2id$v=19$m=...,t=...,p=...$salt$hash.
func HashArgon2id(password []byte) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("passwordhash: generate salt: %w", err)
	}
	kdfInvocations.Add(1)
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
	if err := validateArgon2idParams(m, t, par); err != nil {
		return argon2idParams{}, nil, nil, err
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

// validateArgon2idParams rejects an m/t/p triple parsed from a stored PHC
// string before it is cast to the unsigned wire types and handed to
// argon2.IDKey. Sscanf's %d parses signed integers with no bounds check, so
// a malicious or corrupted hash can encode a negative m/t/p that would
// otherwise wrap to a huge unsigned value on cast (e.g. m=-1 -> a ~4 TiB
// allocation request, an unrecoverable Go runtime OOM). No-fallback
// principle: any violation here is a hard rejection (ErrMalformedHash), never
// a clamp or substituted default.
func validateArgon2idParams(m, t, p int) error {
	if p < minArgon2idParallelism || p > maxArgon2idParallelism {
		return fmt.Errorf("%w: parallelism %d out of range [%d,%d]", ErrMalformedHash, p, minArgon2idParallelism, maxArgon2idParallelism)
	}
	if m < 8*p || m > maxArgon2idMemoryKiB {
		return fmt.Errorf("%w: memory %d out of range [%d,%d]", ErrMalformedHash, m, 8*p, maxArgon2idMemoryKiB)
	}
	if t < minArgon2idTime || t > maxArgon2idTime {
		return fmt.Errorf("%w: time %d out of range [%d,%d]", ErrMalformedHash, t, minArgon2idTime, maxArgon2idTime)
	}
	return nil
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
	kdfInvocations.Add(1)
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
		kdfInvocations.Add(1)
		return bcrypt.CompareHashAndPassword(hash, password) == nil, nil
	case AlgoArgon2id:
		return VerifyArgon2id(string(hash), password)
	default:
		return false, fmt.Errorf("%w: %q", ErrUnrecognizedAlgorithm, algo)
	}
}
