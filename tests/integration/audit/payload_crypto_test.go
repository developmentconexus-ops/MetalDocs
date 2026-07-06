//go:build integration

// payload_crypto_test.go — M7 F7.3 Task C: drives the audit writer's payload
// envelope encryption against a real Postgres instance (testdb factory).
// Proves RecordTx seals payloads under the event's tenant DEK when one
// exists, ListEvents decrypts for display, a destroyed key yields the
// redacted marker (not an error), the hash chain stays green across a mix of
// plaintext and ciphertext events, and the nil-crypto path is byte-identical
// to legacy behavior.
//
//	go test -tags=integration ./tests/integration/audit/... -run PayloadCrypto
package audit_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	securityapp "metaldocs/internal/modules/security/application"
	securitydomain "metaldocs/internal/modules/security/domain"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	platcrypto "metaldocs/internal/platform/crypto"
	"metaldocs/internal/platform/db"
	platformtenant "metaldocs/internal/platform/tenant"
	"metaldocs/tests/integration/testdb"
)

// auditPayloadCryptoAdapter mirrors the composition-root adapter in
// apps/api/cmd/metaldocs-api/main.go: it maps security's typed sentinel
// errors to auditpg.PayloadCrypto's (_, encrypted=false, nil) fall-through
// contract so this test exercises the exact same seam shape production
// wires, without importing the main package.
type auditPayloadCryptoAdapter struct {
	crypto securitydomain.TenantCrypto
}

func (a auditPayloadCryptoAdapter) EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenant(ctx, tenantID, plaintext)
	if err != nil {
		if isSentinel(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return envelope, true, nil
}

func (a auditPayloadCryptoAdapter) DecryptForTenant(ctx context.Context, tenantID, envelope string) ([]byte, error) {
	return a.crypto.DecryptForTenant(ctx, tenantID, envelope)
}

func isSentinel(err error) bool {
	return err == securitydomain.ErrKeyNotFound || err == securitydomain.ErrKeyDestroyed ||
		isWrapped(err, securitydomain.ErrKeyNotFound) || isWrapped(err, securitydomain.ErrKeyDestroyed)
}

func isWrapped(err, target error) bool {
	type unwrapper interface{ Unwrap() error }
	for err != nil {
		if err == target {
			return true
		}
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

func mustKEK(t *testing.T) []byte {
	t.Helper()
	kek := make([]byte, platcrypto.KeySize)
	if _, err := rand.Read(kek); err != nil {
		t.Fatalf("rand.Read kek: %v", err)
	}
	return kek
}

func newCryptoWriter(t *testing.T, sqlDB *sql.DB) (*auditpg.Writer, *securityapp.TenantCryptoService) {
	t.Helper()
	svc, err := securityapp.NewTenantCryptoService(securitypg.NewTenantKeyRepository(sqlDB), mustKEK(t))
	if err != nil {
		t.Fatalf("NewTenantCryptoService() error = %v", err)
	}
	writer := auditpg.NewWriter(sqlDB).WithPayloadCrypto(auditPayloadCryptoAdapter{crypto: svc})
	return writer, svc
}

func ctxForTenant(tenantID string) context.Context {
	ctx := platformtenant.WithTenantID(context.Background(), tenantID)
	ctx = platformtenant.WithActorID(ctx, "test-actor")
	return ctx
}

func newEvent(tenantID, action, payload string) auditdomain.Event {
	return auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      "test-actor",
		Action:       action,
		ResourceType: "audit_test",
		ResourceID:   uuid.NewString(),
		PayloadJSON:  payload,
		TraceID:      "trace-" + uuid.NewString(),
		TenantID:     tenantID,
	}
}

// rawPayload reads the raw stored payload column (bypassing the Writer's
// decrypt-on-read) so tests can assert on the actual on-disk representation.
func rawPayload(t *testing.T, sqlDB *sql.DB, eventID string) string {
	t.Helper()
	var payload string
	if err := sqlDB.QueryRowContext(context.Background(),
		`SELECT payload::text FROM metaldocs.audit_events WHERE id = $1`, eventID,
	).Scan(&payload); err != nil {
		t.Fatalf("read raw payload for %s: %v", eventID, err)
	}
	return payload
}

// (a) KEK-configured + provisioned tenant key: RecordTx stores an envelope,
// ListEvents (the reader) returns the original plaintext.
func TestAuditPayloadCrypto_EncryptsAndDecrypts_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	writer, svc := newCryptoWriter(t, sqlDB)
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	event := newEvent(tenant.ID, "audit.crypto.test", `{"secret":"classified"}`)
	if err := writer.Record(context.Background(), event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	stored := rawPayload(t, sqlDB, event.ID)
	if !platcrypto.IsEnvelope(stored) {
		t.Fatalf("stored payload is not an envelope: %s", stored)
	}

	items, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: event.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListEvents() returned %d items, want 1", len(items))
	}
	if items[0].PayloadJSON != event.PayloadJSON {
		t.Fatalf("decrypted payload = %q, want %q", items[0].PayloadJSON, event.PayloadJSON)
	}
}

// (b) destroy the key -> reader returns the redacted marker, not an error.
func TestAuditPayloadCrypto_DestroyedKey_ReaderRedacts_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	writer, svc := newCryptoWriter(t, sqlDB)
	runner := db.NewTxRunner(sqlDB)
	ctx := ctxForTenant(tenant.ID)

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	event := newEvent(tenant.ID, "audit.crypto.shred_test", `{"secret":"shred me"}`)
	if err := writer.Record(context.Background(), event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	if err := runner.Do(ctx, func(tx *sql.Tx) error {
		return svc.DestroyTenantKeyTx(ctx, tx, tenant.ID)
	}); err != nil {
		t.Fatalf("DestroyTenantKeyTx() error = %v", err)
	}

	items, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: event.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListEvents() returned %d items, want 1", len(items))
	}
	if items[0].PayloadJSON != `{"redacted":"crypto-shredded"}` {
		t.Fatalf("payload after shred = %q, want redacted marker", items[0].PayloadJSON)
	}

	// A NEW event recorded for the now-shredded tenant is a lifecycle
	// tombstone (spec item 6.5) — it must fall through to plaintext, not
	// error out or silently vanish.
	tombstone := newEvent(tenant.ID, "tenant.erased", `{"tombstone":true}`)
	if err := writer.Record(context.Background(), tombstone); err != nil {
		t.Fatalf("Record() tombstone after shred error = %v", err)
	}
	stored := rawPayload(t, sqlDB, tombstone.ID)
	if platcrypto.IsEnvelope(stored) {
		t.Fatalf("tombstone event stored as envelope, want plaintext: %s", stored)
	}
	// payload is a jsonb column — Postgres canonicalizes spacing on write, so
	// the contract is semantic identity, not byte identity.
	if !jsonSemanticallyEqual(t, stored, tombstone.PayloadJSON) {
		t.Fatalf("tombstone stored payload = %q, want semantically %q", stored, tombstone.PayloadJSON)
	}
}

// jsonSemanticallyEqual compares two JSON documents by value (the payload
// column is jsonb: Postgres rewrites whitespace/key-order, making byte
// comparison meaningless for plaintext payloads).
func jsonSemanticallyEqual(t *testing.T, a, b string) bool {
	t.Helper()
	var va, vb any
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		t.Fatalf("unmarshal %q: %v", a, err)
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		t.Fatalf("unmarshal %q: %v", b, err)
	}
	return reflect.DeepEqual(va, vb)
}

// (c) chain integrity across a mix of plaintext (no-key tenant) and
// ciphertext (keyed tenant) events: ValidateIntegrity reports 0 issues.
func TestAuditPayloadCrypto_MixedChainStaysGreen_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	keyedTenant := testdb.NewTenant(t, sqlDB)
	plainTenant := testdb.NewTenant(t, sqlDB)
	writer, svc := newCryptoWriter(t, sqlDB)
	runner := db.NewTxRunner(sqlDB)

	if err := runner.Do(ctxForTenant(keyedTenant.ID), func(tx *sql.Tx) error {
		return svc.ProvisionTenantKeyTx(ctxForTenant(keyedTenant.ID), tx, keyedTenant.ID)
	}); err != nil {
		t.Fatalf("ProvisionTenantKeyTx() error = %v", err)
	}

	// plainTenant never gets a key -> stays legacy plaintext.
	for i := 0; i < 3; i++ {
		if err := writer.Record(context.Background(), newEvent(plainTenant.ID, "audit.plain", `{"n":1}`)); err != nil {
			t.Fatalf("Record() plaintext event error = %v", err)
		}
		if err := writer.Record(context.Background(), newEvent(keyedTenant.ID, "audit.cipher", `{"n":2}`)); err != nil {
			t.Fatalf("Record() ciphertext event error = %v", err)
		}
	}

	issues, err := writer.ValidateIntegrity(context.Background())
	if err != nil {
		t.Fatalf("ValidateIntegrity() error = %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("ValidateIntegrity() issues = %#v, want none", issues)
	}
}

// (d) nil encryptor: byte-identical plaintext path (today's legacy
// behavior), proving WithPayloadCrypto is fully opt-in.
func TestAuditPayloadCrypto_NilCrypto_PlaintextByteIdentical_RealDB(t *testing.T) {
	sqlDB, _ := testdb.Open(t)
	tenant := testdb.NewTenant(t, sqlDB)
	writer := auditpg.NewWriter(sqlDB) // no WithPayloadCrypto call

	event := newEvent(tenant.ID, "audit.legacy", `{"legacy":true}`)
	if err := writer.Record(context.Background(), event); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	stored := rawPayload(t, sqlDB, event.ID)
	// jsonb canonicalizes spacing — semantic identity is the contract.
	if !jsonSemanticallyEqual(t, stored, event.PayloadJSON) {
		t.Fatalf("stored payload = %q, want semantically %q", stored, event.PayloadJSON)
	}

	items, _, err := writer.ListEvents(context.Background(), auditdomain.ListEventsQuery{
		TenantID: tenant.ID, ResourceID: event.ResourceID, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(items) != 1 || !jsonSemanticallyEqual(t, items[0].PayloadJSON, event.PayloadJSON) {
		t.Fatalf("ListEvents() payload = %+v, want semantically %q", items, event.PayloadJSON)
	}
}
