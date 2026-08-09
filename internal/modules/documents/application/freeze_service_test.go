package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
)

// fakeTx is a no-op db.Tx for unit tests that don't need real SQL execution.
type fakeTx struct{}

func (fakeTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, nil
}
func (fakeTx) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, nil
}
func (fakeTx) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}

type fakeSnapshotReader struct {
	snap           v2dom.TemplateSnapshot
	valuesFrozenAt *time.Time
	err            error
	// currentRef is the document's head editor revision — what Pin pins.
	// frozenRef is the revision the pin NAMES — what Materialize renders.
	// They are separate fields on purpose: tests can make them diverge and
	// prove Materialize follows the pin instead of the head.
	currentRef v2dom.RevisionRef
	frozenRef  v2dom.RevisionRef
}

func (f fakeSnapshotReader) ReadCurrentRevisionRef(_ context.Context, _, _ string, _ ...infrastructure.DBTX) (v2dom.RevisionRef, error) {
	return f.currentRef, f.err
}

func (f fakeSnapshotReader) ReadFrozenRevisionRef(_ context.Context, _, _ string, _ ...infrastructure.DBTX) (v2dom.RevisionRef, error) {
	return f.frozenRef, f.err
}

// fakeRevisionBodyReader serves bytes per storage key. A key that is absent
// returns an error, so "Materialize read a key nobody staged" surfaces as a
// failure instead of an empty body that would hash to the empty-string digest.
type fakeRevisionBodyReader struct {
	bodies map[string][]byte
	keys   []string
	err    error
}

func (f *fakeRevisionBodyReader) AssertedReadObject(_ context.Context, _, key string, _ int64) ([]byte, error) {
	f.keys = append(f.keys, key)
	if f.err != nil {
		return nil, f.err
	}
	body, ok := f.bodies[key]
	if !ok {
		return nil, fmt.Errorf("fake body reader: no object at %s", key)
	}
	return body, nil
}

// hashOf is the digest document_revisions.content_hash records for a body.
func hashOf(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func (f fakeSnapshotReader) ReadSnapshotWithFreezeAt(_ context.Context, _, _ string, _ ...infrastructure.DBTX) (v2dom.TemplateSnapshot, *time.Time, error) {
	return f.snap, f.valuesFrozenAt, f.err
}

func (f fakeSnapshotReader) ReadFreezeAt(_ context.Context, _, _ string, _ ...infrastructure.DBTX) (*time.Time, error) {
	return f.valuesFrozenAt, f.err
}

type fakeFanoutClient struct {
	req   fanout.Request
	resp  fanout.Response
	err   error
	calls int
}

func (f *fakeFanoutClient) Fanout(_ context.Context, req fanout.Request) (fanout.Response, error) {
	f.calls++
	f.req = req
	if f.err != nil {
		return fanout.Response{}, f.err
	}
	return f.resp, nil
}

type fakeValuesReader struct {
	values []infrastructure.PlaceholderValue
	err    error
	calls  int
}

func (f *fakeValuesReader) ListValues(_ context.Context, _, _ string) ([]infrastructure.PlaceholderValue, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.values, nil
}

type fakeFreezeFinalizer struct {
	calls int
	hash  []byte
	at    time.Time
	err   error
	// pinnedRevisionID captures the document_revisions.id Pin stamped, so
	// tests can assert the lineage pin is written and is NOT the documents.id
	// the pipeline's `revisionID` parameter carries.
	pinnedRevisionID string
}

func (f *fakeFreezeFinalizer) WriteFreeze(_ context.Context, _, _ string, valuesHash []byte, frozenRevisionID string, frozenAt time.Time, _ ...infrastructure.DBTX) error {
	if f.err != nil {
		return f.err
	}
	f.calls++
	f.hash = append([]byte(nil), valuesHash...)
	f.pinnedRevisionID = frozenRevisionID
	f.at = frozenAt
	return nil
}

type fakeResolverContextBuilder struct {
	input resolvers.ResolveInput
	err   error
	calls int
}

func (f *fakeResolverContextBuilder) Build(_ context.Context, _, _ string, _ ApproverContext) (resolvers.ResolveInput, error) {
	if f.err != nil {
		return resolvers.ResolveInput{}, f.err
	}
	f.calls++
	return f.input, nil
}

func (f *fakeResolverContextBuilder) BuildForDraft(_ context.Context, _, _ string) (resolvers.ResolveInput, error) {
	if f.err != nil {
		return resolvers.ResolveInput{}, f.err
	}
	return f.input, nil
}

type fixedResolver struct {
	key string
	ver int
	val any
}

func (r fixedResolver) Key() string  { return r.key }
func (r fixedResolver) Version() int { return r.ver }
func (r fixedResolver) Resolve(_ context.Context, _ resolvers.ResolveInput) (resolvers.ResolvedValue, error) {
	return resolvers.ResolvedValue{
		Value:       r.val,
		ResolverKey: r.key,
		ResolverVer: r.ver,
		InputsHash:  []byte{0xaa},
		ComputedAt:  time.Now().UTC(),
	}, nil
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func strPtr(v string) *string { return &v }

func bytesToHex(b []byte) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexDigits[v>>4]
		out[i*2+1] = hexDigits[v&0x0f]
	}
	return string(out)
}
