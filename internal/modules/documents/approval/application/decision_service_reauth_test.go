package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"metaldocs/internal/modules/documents/approval/infrastructure/signature"
	"metaldocs/internal/modules/documents/approval/repository"
)

// reauthFakeUserReader satisfies signature.IamUserReader for the reauth tests.
type reauthFakeUserReader struct {
	hashes map[string][]byte
}

func newReauthReader(t *testing.T, users map[string]string) reauthFakeUserReader {
	t.Helper()
	hashes := make(map[string][]byte, len(users))
	for id, pw := range users {
		h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("bcrypt hash: %v", err)
		}
		hashes[id] = h
	}
	return reauthFakeUserReader{hashes: hashes}
}

func (r reauthFakeUserReader) GetPasswordHash(_ context.Context, userID string) ([]byte, error) {
	h, ok := r.hashes[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return h, nil
}

// newReauthRegistry wires a real password_reauth provider backed by the fake reader.
func newReauthRegistry(reader signature.IamUserReader) *signature.Registry {
	reg := signature.NewRegistry()
	reg.Register(signature.NewPasswordReauthProvider(
		context.Background(),
		reader,
		nil,
		signature.NewInMemoryAuthFailureRateLimiter(),
	))
	return reg
}

func newReauthDecisionService(reg *signature.Registry, repo repository.ApprovalRepository, emitter EventEmitter) *DecisionService {
	return (&DecisionService{
		repo:          repo,
		emitter:       emitter,
		clock:         fixedClock{t: time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)},
		freezeInvoker: &fakeFreezeInvoker{},
		pdfDispatcher: &fakePDFDispatchInvoker{},
	}).WithSignatureRegistry(reg)
}

func reauthSignoffRequest(instanceID, stageID, actorID, password string) SignoffRequest {
	return SignoffRequest{
		TenantID:         "tenant-1",
		InstanceID:       instanceID,
		StageInstanceID:  stageID,
		ActorUserID:      actorID,
		Decision:         "approve",
		Comment:          "LGTM",
		SignatureMethod:  "password_reauth",
		SignaturePayload: map[string]any{"password_token": password},
		ContentFormData:  map[string]any{"_content_hash": validContentHash},
	}
}

// TestRecordSignoff_Reauth_RejectsBlankToken: a blank password_token must be
// rejected with signature.ErrInvalidCredentials and must NOT record a sign-off.
func TestRecordSignoff_Reauth_RejectsBlankToken(t *testing.T) {
	const (
		instanceID = "inst-reauth-blank"
		stageID    = "stage-reauth-blank"
		actorID    = "approver-blank"
		authorID   = "author-blank"
	)
	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	conn := &decisionTestConn{authzGranted: true, areaCode: "QA", actorID: actorID}
	repo := &fakeDecisionRepo{instance: inst}
	reg := newReauthRegistry(newReauthReader(t, map[string]string{actorID: "correct-pass"}))
	svc := newReauthDecisionService(reg, repo, &MemoryEmitter{})
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), db, reauthSignoffRequest(instanceID, stageID, actorID, ""))
	if !errors.Is(err, signature.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want signature.ErrInvalidCredentials", err)
	}
	if repo.insertedSignoff != nil {
		t.Fatal("sign-off must not be recorded on blank password_token")
	}
}

// TestRecordSignoff_Reauth_RejectsWrongPassword: a wrong password must be
// rejected (disclosure-safe) and must NOT record a sign-off.
func TestRecordSignoff_Reauth_RejectsWrongPassword(t *testing.T) {
	const (
		instanceID = "inst-reauth-wrong"
		stageID    = "stage-reauth-wrong"
		actorID    = "approver-wrong"
		authorID   = "author-wrong"
	)
	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	conn := &decisionTestConn{authzGranted: true, areaCode: "QA", actorID: actorID}
	repo := &fakeDecisionRepo{instance: inst}
	reg := newReauthRegistry(newReauthReader(t, map[string]string{actorID: "correct-pass"}))
	svc := newReauthDecisionService(reg, repo, &MemoryEmitter{})
	db := newDecisionTestDB(t, conn)

	_, err := svc.RecordSignoff(context.Background(), db, reauthSignoffRequest(instanceID, stageID, actorID, "wrong-pass"))
	if !errors.Is(err, signature.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want signature.ErrInvalidCredentials", err)
	}
	if repo.insertedSignoff != nil {
		t.Fatal("sign-off must not be recorded on wrong password")
	}
}

// TestRecordSignoff_Reauth_AcceptsCorrectPassword: the correct password records
// the sign-off and persists the attestation (no raw credential).
func TestRecordSignoff_Reauth_AcceptsCorrectPassword(t *testing.T) {
	const (
		instanceID = "inst-reauth-ok"
		stageID    = "stage-reauth-ok"
		actorID    = "approver-ok"
		authorID   = "author-ok"
	)
	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	signedAt := time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)
	stageSignoffs := []signoffRow{{
		id:                 "signoff-reauth-ok",
		approvalInstanceID: instanceID,
		stageInstanceID:    stageID,
		actorUserID:        actorID,
		actorTenantID:      "tenant-1",
		decision:           "approve",
		signedAt:           signedAt,
		signatureMethod:    "password_reauth",
		signaturePayload:   []byte(`{}`),
		contentHash:        validContentHash,
	}}
	conn := &decisionTestConn{stageSignoffs: stageSignoffs, authzGranted: true, areaCode: "QA", actorID: actorID}
	repo := &fakeDecisionRepo{
		instance:         inst,
		insertSignoffRes: repository.SignoffInsertResult{ID: "signoff-reauth-ok", WasReplay: false},
	}
	reg := newReauthRegistry(newReauthReader(t, map[string]string{actorID: "correct-pass"}))
	svc := newReauthDecisionService(reg, repo, &MemoryEmitter{})
	db := newDecisionTestDB(t, conn)

	result, err := svc.RecordSignoff(context.Background(), db, reauthSignoffRequest(instanceID, stageID, actorID, "correct-pass"))
	if err != nil {
		t.Fatalf("RecordSignoff: unexpected error: %v", err)
	}
	if !result.StageCompleted || !result.InstanceApproved {
		t.Fatalf("result = %+v, want stage completed + instance approved", result)
	}
	if repo.insertedSignoff == nil {
		t.Fatal("expected a recorded sign-off")
	}
	// The persisted signature payload must be the verified attestation, never the
	// raw credential the client supplied.
	payload := string(repo.insertedSignoff.SignaturePayload())
	if strings.Contains(payload, "password_token") || strings.Contains(payload, "correct-pass") {
		t.Fatalf("stored signature payload leaks the raw credential: %s", payload)
	}
	if !strings.Contains(payload, "password_reauth") {
		t.Fatalf("stored signature payload = %s, want the reauth attestation", payload)
	}
}

// TestRecordSignoff_Reauth_RateLimitTrips: repeated bad attempts trip the rate
// limiter (signature.ErrRateLimited) rather than continuing to bcrypt-compare.
func TestRecordSignoff_Reauth_RateLimitTrips(t *testing.T) {
	const (
		instanceID = "inst-reauth-rl"
		stageID    = "stage-reauth-rl"
		actorID    = "approver-rl"
		authorID   = "author-rl"
	)
	inst := buildSingleStageInstance(instanceID, stageID, authorID, []string{actorID})
	reg := newReauthRegistry(newReauthReader(t, map[string]string{actorID: "correct-pass"}))
	repo := &fakeDecisionRepo{instance: inst}
	svc := newReauthDecisionService(reg, repo, &MemoryEmitter{})

	for i := 0; i < 5; i++ {
		conn := &decisionTestConn{authzGranted: true, areaCode: "QA", actorID: actorID}
		db := newDecisionTestDB(t, conn)
		_, err := svc.RecordSignoff(context.Background(), db, reauthSignoffRequest(instanceID, stageID, actorID, "wrong-pass"))
		if !errors.Is(err, signature.ErrInvalidCredentials) {
			t.Fatalf("attempt %d err = %v, want signature.ErrInvalidCredentials", i, err)
		}
	}
	conn := &decisionTestConn{authzGranted: true, areaCode: "QA", actorID: actorID}
	db := newDecisionTestDB(t, conn)
	_, err := svc.RecordSignoff(context.Background(), db, reauthSignoffRequest(instanceID, stageID, actorID, "wrong-pass"))
	if !errors.Is(err, signature.ErrRateLimited) {
		t.Fatalf("err = %v, want signature.ErrRateLimited after 5 failures", err)
	}
}
