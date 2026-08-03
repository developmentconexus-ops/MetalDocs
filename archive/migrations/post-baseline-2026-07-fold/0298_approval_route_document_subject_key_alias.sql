-- 0298_approval_route_document_subject_key_alias.sql
-- ROADMAP M3 P3.S2b-2 regression fix (approval kernel extraction). Prior
-- commit c0ae114a made LoadActiveRouteIDByProfile delegate to
-- LoadActiveRouteIDBySubject(tenantID, "document", profileCode) — i.e. it now
-- looks up a document route by (subject_kind='document', subject_key=
-- profile_code). Per ratified rail R1, profile_code is the backward-compat
-- ALIAS for (document, profile_code): that delegation is only correct if
-- EVERY document route's subject_key equals its profile_code.
--
-- That was not guaranteed. resolveCreateRouteSubject
-- (internal/modules/approval/application/route_admin_service.go) defaulted
-- subject_key to profile_code only when the client sent an empty key; an
-- explicit subject_kind=document plus a divergent subject_key was accepted
-- as-is (proven by http/route_admin_handler_test.go
-- TestCreateRoute_SubjectFieldsPassedThrough, which pre-fix created a
-- profile_code="ops"/subject_kind="document"/subject_key="custom-key" route).
-- Such a route becomes unfindable by the profile-code alias lookup, so
-- document submission would regress to ErrApprovalRouteMissing.
--
-- This migration is the DB-level (last-line) half of the fix: an app-level
-- guard in resolveCreateRouteSubject now rejects a divergent document
-- subject_key at create time (friendly first line); this migration adds the
-- matching CHECK constraint so the DB enforces the same invariant as the
-- authoritative last line, per this project's "DB enforces invariants" rule.
--
-- EXPAND phase only, forward-only, idempotent — same idiom as 0297
-- (DROP CONSTRAINT IF EXISTS before ADD CONSTRAINT; ledger row is
-- ON CONFLICT DO NOTHING).
--
--   1. Normalize any pre-existing divergent dev rows so the CHECK below can
--      be added without failing on existing data. profile_code is guaranteed
--      NOT NULL for every document-subject row by 0297's
--      approval_routes_document_subject_projection_check, so this UPDATE
--      always has a valid, non-NULL value to fold subject_key onto.
--
--   2. Add approval_routes_document_subject_key_alias_check: for a
--      document-subject row, subject_key must equal profile_code. This sits
--      alongside 0297's projection CHECK constraints (which govern
--      subject_kind vs. column presence), not replacing them — the alias
--      check governs subject_kind vs. subject_key VALUE for the document
--      case specifically. Template rows are unconstrained by this check
--      (subject_kind <> 'document' short-circuits it), matching the app-level
--      guard which only fires for subject_kind=document.
--
-- Schema + dev-data-normalize only: no Go wiring to the new constraint name
-- is needed (Postgres enforces it automatically on every INSERT/UPDATE of
-- approval_routes going forward; the app-level guard in
-- resolveCreateRouteSubject is what turns a violation into a friendly 4xx
-- before it ever reaches this constraint).
--
-- Multi-tenant: no new table, index, or column — nothing to tenant-scope.
-- RLS FORCE on approval_routes is untouched (unaffected by adding a CHECK).
--
-- Out of scope: the UPDATE/supersede path in route_admin_service.go carries
-- subject_kind/subject_key forward UNCHANGED from the existing row (INSERT
-- ... SELECT subject_kind, subject_key FROM approval_routes WHERE id = $4),
-- so it cannot introduce a new divergence and needs no additional guard.

BEGIN;

-- 1. Normalize any pre-existing divergent dev rows ---------------------------

UPDATE public.approval_routes
   SET subject_key = profile_code
 WHERE subject_kind = 'document'
   AND subject_key IS DISTINCT FROM profile_code;

-- 2. Document-subject alias CHECK constraint ---------------------------------

ALTER TABLE public.approval_routes
    DROP CONSTRAINT IF EXISTS approval_routes_document_subject_key_alias_check;
ALTER TABLE public.approval_routes
    ADD CONSTRAINT approval_routes_document_subject_key_alias_check
    CHECK (subject_kind <> 'document' OR subject_key = profile_code);

-- ── schema_migrations ledger ─────────────────────────────────────────────────────────────

INSERT INTO public.schema_migrations (version, description)
VALUES ('0298', 'M3 P3.S2b-2 regression fix: normalize pre-existing divergent approval_routes dev rows (subject_key := profile_code where subject_kind=document) and add approval_routes_document_subject_key_alias_check CHECK (subject_kind <> ''document'' OR subject_key = profile_code) so the DB enforces the document-route profile_code alias invariant as the last line, matching the new app-level guard in resolveCreateRouteSubject (ErrDocumentSubjectKeyMismatch).')
ON CONFLICT (version) DO NOTHING;

COMMIT;
