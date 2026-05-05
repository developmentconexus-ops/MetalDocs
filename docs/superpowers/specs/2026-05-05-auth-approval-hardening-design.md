# Sub-project A — Auth & Approval Hardening — Design

> **Date:** 2026-05-05
> **Status:** Design (pre-implementation)
> **Scope:** Fix bugs **J2** (permissive authz), **J1** (signoff eligibility not enforced), **M1** (fanout client no timeout) from `wiki/bugs/audit-2026-05-04.md`.
> **Out of scope:** Sub-project B (frontend v1 purge), Sub-project C (dead code), broader refactor of `decision_service.go`.

## Context

Pre-release security/correctness audit (2026-05-04) found three issues that block any production-bound deploy:

- **J2 (critical):** `permissiveAuthzChecker` (`apps/api/cmd/metaldocs-api/main.go:482`) returns `nil` for every check. Wired into documents module via `AuthorizationChecker` interface. Effective scope: gates `iamdomain.CapDocumentCreate` only (one caller at `internal/modules/documents/application/service.go:240`); other capabilities are protected by `authz.Require` in-tx. Still — a bypass shipped in `main.go` is a hard fail.
- **J1 (critical):** `decision_service.RecordSignoff` (`internal/modules/documents/approval/application/decision_service.go:141`) calls `authz.Require("doc.signoff", areaCode)` and `domain.CheckSoD(...)` but never checks `req.ActorUserID ∈ activeStage.EligibleActorIDs`. Any user with the area-scoped capability can sign — including users not in the snapshot frozen at submit time. Violates ISO 9001 segregation guarantee promised by `eligible_actor_ids` snapshot.
- **M1 (medium):** `fanoutCli = fanout.NewClient(fanoutURL, serviceToken, nil)` (`apps/api/cmd/metaldocs-api/main.go:238`). `nil` falls back to `http.DefaultClient` — no timeout, default `MaxIdleConnsPerHost=2`. If Gotenberg hangs, goroutines leak and connections pile up.

## Design principles applied

- ADR 0007 two-tier-authz boundary preserved.
- Domain invariants live in `domain/` pkg as pure functions (mirrors existing `sod.go`, `quorum.go`, `drift.go`).
- Defense in depth on regulated signoff path: app check + DB constraint + audit log.
- Reusable platform primitives lifted into `internal/platform/`.
- No new observability infra (Prometheus/OTel) — defer until metrics stack lands.
- No retry/circuit breaker in fanout client — `PDFOutboxWorker` (ADR 0009) owns retry.

---

## J2 — Replace `permissiveAuthzChecker`

### Decision

Delete the `AuthorizationChecker` interface owned by the documents module. Documents module declares a 1-method consumer port matching `iam/application.CapabilityService.CanDo`; production wiring binds the port directly to `*iamapp.CapabilityService`. Documents-module wiring is lifted out of `main.go` into a dedicated wiring file.

### Why this fits MetalDocs

- ADR 0007 places `document.create` at tier 1 (tenant-level capability, no area chosen at creation time). `CapabilityService.CanDo` is the canonical tier-1 check.
- `AuthorizationChecker` and `CapabilityService.CanDo` have identical semantics (`(ctx, userID, tenantID, capability) → error`). The adapter would be dead weight.
- Lifting wiring out of `main.go` (already 500+ lines) addresses god-file drift and establishes the pattern for future module-wiring extractions.

### Changes

**Backend (`internal/modules/documents/application/`):**
- `service.go:89-92` — remove `AuthorizationChecker` interface and `Check` method definition.
- `service.go:240` — call site stays the same shape: `s.caps.CanDo(ctx, actorID, tenantID, iamdomain.CapDocumentCreate)`.
- `service.go` `Dependencies` struct — replace `AuthzChecker AuthorizationChecker` with `CapabilityChecker CapabilityChecker` (port type below).
- New file `internal/modules/documents/application/ports.go` — declare consumer port:
  ```go
  type CapabilityChecker interface {
      CanDo(ctx context.Context, userID, tenantID string, cap iamdomain.Capability) error
  }
  ```
  (Documents module owns this declaration of *what it needs*. IAM module's concrete `CapabilityService` satisfies it structurally.)

**Wiring (`apps/api/cmd/metaldocs-api/main.go` and new file):**
- New `apps/api/internal/wiring/documents.go` — exports `WireDocuments(deps WireDeps) documents.Dependencies` constructing the documents-module `Dependencies` struct, wiring `CapabilityChecker: capabilityService`.
- `main.go` — remove `permissiveAuthzChecker` struct (lines 482-487). Replace inline `documents.Dependencies{...}` construction with `wiring.WireDocuments(...)` call.

### Tests

- Existing documents-module tests passing without `permissiveAuthzChecker` test stub.
- Add unit test asserting `service.CreateDocument` returns `iam.ErrCapabilityDenied` when `CapabilityChecker` rejects.

---

## J1 — Enforce signoff eligibility (defense in depth)

### Decision

Add eligibility as a pure domain invariant. Enforce at three layers:

1. **Domain (pure function):** `domain.CheckEligibility(actorUserID, eligibleActorIDs)` returning `ErrActorNotEligible`.
2. **Service:** `decision_service.RecordSignoff` calls it inside the existing tx, after `authz.Require`, before `CheckSoD`. Emits audit event on rejection.
3. **DB (defense in depth):** trigger on `approval_signoffs` INSERT verifies actor ∈ parent stage's `eligible_actor_ids` jsonb. Migration `0180_signoff_eligibility_trigger.sql`.

### Why this fits MetalDocs

- Domain folder already organises one rule per file (`sod.go`, `quorum.go`, `drift.go`, `route.go`, `signoff.go`, `state.go`). `eligibility.go` slots in.
- Pure function = trivially testable, mirrors `sod.go` shape (no DB, no mocks).
- DB trigger pattern matches existing `enforce_document_transition` (migration 0142) — established repo convention for regulated state guards.
- Audit event uses existing `audit.Recorder` already injected in `decision_service` — no new infrastructure.

### Changes

**Domain — new file `internal/modules/documents/approval/domain/eligibility.go`:**
```go
package domain

import "errors"

var ErrActorNotEligible = errors.New("eligibility: actor is not in the eligible_actor_ids snapshot for the active stage")

// CheckEligibility verifies actorUserID is present in the eligible-actor snapshot
// frozen at submit time. Pure function — no DB, no globals.
func CheckEligibility(actorUserID string, eligibleActorIDs []string) error {
    for _, id := range eligibleActorIDs {
        if id == actorUserID {
            return nil
        }
    }
    return ErrActorNotEligible
}
```

**Domain test — new file `eligibility_test.go`:**
- Example test: actor in / not in / empty list / single-element list.
- Property test: for any random eligible set, function rejects iff actor ∉ set (using `testing/quick` or table-driven random gen).

**Service — `decision_service.go`:**
- After `authz.Require(...)` at line 141, before `CheckSoD` block at line 158:
  ```go
  if err := domain.CheckEligibility(req.ActorUserID, activeStage.EligibleActorIDs); err != nil {
      _ = auditRec.Record(ctx, audit.Event{
          Kind:        "signoff.rejected",
          ActorUserID: req.ActorUserID,
          ResourceID:  req.InstanceID,
          Reason:      "not_eligible",
      })
      return err
  }
  ```
- Verify `repository.LoadActiveStageForUpdate` (or equivalent) loads the stage row with `FOR UPDATE`. If not, add it. Prevents race against concurrent re-snapshot via re-submit.

**Repository — confirm `FOR UPDATE` lock:**
- Read `internal/modules/documents/approval/repository/postgres_approval_repository.go` `LoadActiveStage` SQL.
- If missing `FOR UPDATE`, add it. Document in inline comment.

**HTTP handler — error mapping:**
- `internal/modules/documents/approval/http/...` — map `domain.ErrActorNotEligible` to HTTP 403 with code `signoff.not_eligible` (per `wiki/concepts/error-ux.md` E2 dialog states pattern).

**Migration — `migrations/0180_signoff_eligibility_trigger.sql`:**
```sql
CREATE OR REPLACE FUNCTION enforce_signoff_eligibility() RETURNS trigger AS $$
DECLARE
    eligible jsonb;
BEGIN
    SELECT eligible_actor_ids INTO eligible
    FROM   approval_stage_instances
    WHERE  id = NEW.stage_instance_id;

    IF eligible IS NULL OR NOT eligible @> to_jsonb(NEW.actor_user_id::text) THEN
        RAISE EXCEPTION 'signoff: actor % is not in eligible_actor_ids for stage %',
            NEW.actor_user_id, NEW.stage_instance_id
            USING ERRCODE = '23514';  -- check_violation
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_signoff_eligibility_trg
BEFORE INSERT ON approval_signoffs
FOR EACH ROW EXECUTE FUNCTION enforce_signoff_eligibility();

INSERT INTO public.schema_migrations (version, description)
VALUES ('0180', 'enforce signoff eligibility against eligible_actor_ids snapshot')
ON CONFLICT (version) DO NOTHING;
```

### Tests

- `eligibility_test.go` — examples + property test.
- Integration test: signoff with non-eligible actor returns 403 + `audit_events` row written.
- Migration test: direct SQL INSERT into `approval_signoffs` with non-eligible actor raises trigger error (defense in depth verified).

---

## M1 — Fanout client tuned transport

### Decision

New `internal/platform/httpclient/internal_client.go` exporting `NewInternalClient()` with explicit transport tuning. Wire fanout client to use it. Per-call context deadline at dispatch site. Inline comment documents retry ownership.

### Why this fits MetalDocs

- Sets reusable defaults for the next service-to-service client (no copy-paste).
- Tuned `http.Transport` is the canonical Go pattern; `http.Client{Timeout: 30s}` alone misses connect/handshake/header timeouts and uses anemic default pool sizes.
- No new infra: no metrics, no tracing — pure stdlib config.
- Outbox pattern (ADR 0009) already owns retry; client must fail fast, not double-retry.

### Changes

**New file `internal/platform/httpclient/internal_client.go`:**
```go
package httpclient

import (
    "net"
    "net/http"
    "time"
)

// NewInternalClient returns an *http.Client with tuned defaults for
// service-to-service calls inside the MetalDocs deployment (API → fanout,
// API → future internal services).
//
// Retry/backoff is intentionally absent: callers either fail-fast and rely
// on a transactional outbox for retry (see ADR 0009), or own retry at
// their layer.
func NewInternalClient() *http.Client {
    transport := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   5 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        IdleConnTimeout:       90 * time.Second,
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   20,
        MaxConnsPerHost:       50,
        ForceAttemptHTTP2:     true,
    }
    return &http.Client{
        Transport: transport,
        Timeout:   60 * time.Second,
    }
}
```

**Wiring — `apps/api/cmd/metaldocs-api/main.go`:**
- Import new package.
- Replace `fanoutCli = fanout.NewClient(fanoutURL, serviceToken, nil)` with `fanoutCli = fanout.NewClient(fanoutURL, serviceToken, httpclient.NewInternalClient())`.

**Dispatch site — `internal/modules/documents/.../pdf_dispatcher.go` (or fanout client itself):**
- Wrap outbound call with per-call deadline:
  ```go
  callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
  defer cancel()
  resp, err := c.do(req.WithContext(callCtx))
  ```
- Add inline comment: `// Client fails fast; retry owned by PDFOutboxWorker (ADR 0009).`

### Tests

- Unit test for `NewInternalClient` — asserts transport timeouts present (table-driven on `http.Transport` exported fields).
- Integration check (manual): kill Gotenberg mid-render, observe API does not leak goroutines, outbox row stays for retry.

---

## Cross-cutting

### Migrations
- `migrations/0180_signoff_eligibility_trigger.sql` — single new migration.

### Wiki updates (defer to wiki-curator after implementation)
- `wiki/modules/approval.md` — add eligibility check to RecordSignoff flow, link to new domain file.
- `wiki/decisions/0007-two-tier-authz.md` — note `document.create` wiring path uses CapabilityChecker port; permissiveAuthzChecker removed.
- `wiki/bugs/audit-2026-05-04.md` — mark J2/J1/M1 resolved.
- `wiki/architecture/system-overview.md` — note `internal/platform/httpclient` package.

### ADR updates
- ADR 0007 amendment: short note that `document.create` is wired via consumer port, no `permissiveAuthzChecker`.
- ADR 0009 — no amendment (inline code comment is sufficient).

### Out-of-scope follow-ups (logged, not done here)
- `decision_service.go` is 499 LOC — borderline god-file. Future "Sub-project D: approval module hygiene" splits by command (per-action files) or CQRS query/command split. Not part of this spec.
- ADR README cleanup (status-field updates for 0003, vendor-path fix for 0001) — separate `wiki: ADR refresh` task to dispatch via wiki-curator.

---

## Acceptance criteria

- [ ] `permissiveAuthzChecker` deleted from `main.go`. `documents.Dependencies` takes `CapabilityChecker` port. Wiring lives in `apps/api/internal/wiring/documents.go`.
- [ ] `CreateDocument` returns `iam.ErrCapabilityDenied` for users without `CapDocumentCreate` (integration test green).
- [ ] `domain/eligibility.go` exists with pure `CheckEligibility` + property test.
- [ ] `RecordSignoff` rejects non-eligible actor with `domain.ErrActorNotEligible`, HTTP 403, audit event `signoff.rejected{reason=not_eligible}`.
- [ ] Stage row loaded `FOR UPDATE` during signoff.
- [ ] Migration `0180_signoff_eligibility_trigger.sql` rejects raw SQL INSERT bypassing app layer (defense-in-depth verified).
- [ ] `internal/platform/httpclient/internal_client.go` exists. Fanout client uses it. Per-call `context.WithTimeout` at dispatch.
- [ ] Inline comment at dispatch site documents retry ownership (ADR 0009).
- [ ] All existing tests pass; new tests added per section.
