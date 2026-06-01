# Approval Full-Flow QA Run — 2026-06-01

**Verdict:** PASS-WITH-WARNINGS — happy-path signoff + content-pin + OCC + governance events confirmed; **one CRITICAL regression** found on the taxonomy authz cap downgrade (commit `8e1518c85`); one MEDIUM idempotency-replay semantics gap on doc-scoped signoff.

## 1. Environment

| Item | Value |
|---|---|
| Branch | `main` |
| Commit SHA | `8e1518c85aa2f15ac1f1879bd3abaf69dac682cf` |
| Startup command | `.\scripts\start-api.ps1 -Build` |
| API port | `:8081` (binary rebuilt; migrations 14/14 applied) |
| Workers | `metaldocs-worker` PID 14712, `metaldocs-jobs` PID 40840 |
| docx-renderer | `:3100` reachable (HEAD `/healthz`=404, `/`=307 — service up) |
| Auth scheme | Session cookie (`metaldocs_session`); `SameSite=Strict` + `Origin` enforcement required |
| Login (admin) | 200, `roles=[system_admin]`, caps include `document.signoff` + `taxonomy.manage` + `document.view` |
| Login (approver) | 200, `roles=[approver]`, caps `document.{create,edit,signoff,submit,view}` + `template.{approve,review,view}` — **no** `taxonomy.manage` (key for regression test) |
| Login (approver-test) | 401 `AUTH_INVALID_CREDENTIALS` — password drift (see Defers) |

## 2. Test transcript

All mutating requests sent with `Origin: http://localhost:8081` (`SameSite=Strict` blocks otherwise → 403 `AUTH_INVALID_ORIGIN`).

Target instance: `55e699f7-6609-4a2b-ac5d-ad5d988851b1` (doc `c1bb2112-21ea-46fc-ac1f-719d04994d41`, CD `979d15f9-…`, code `PO-RH-002`, status `under_review`, stage `Approval` quorum `0/1`, eligible `[approver, approver-test]`, submitter `admin`).

### 2.1 Read path (approver)

| Step | Request | Expected | Actual | Result |
|---|---|---|---|---|
| Inbox | `GET /api/v1/approval/inbox` | 200, item for doc | `200 {"items":[{"instance_id":"55e699f7-…","document_id":"c1bb2112-…","quorum_progress":"0/1"}],"total":1}` | ✓ |
| Active-document | `GET /api/v1/controlled-documents/979d15f9-…/active-document` | 200 with `contentHash` | `200 {"contentHash":"a66f61a80f8e3894ab956700229ac0f5efe6f8a4cc01a9ec78faba95f55d71a3","revisionVersion":1,"approvalState":"under_review"}` | ✓ |
| Get instance | `GET /api/v1/documents/c1bb2112-…/approval-instance` | 200, stage 1 active | `200`, `status=in_progress`, `etag="v1"`, stage active | ✓ |

### 2.2 Negative path — content-pin, OCC, preconditions

Idempotency-Key unique per call (`qa-<tag>-<nanos>`). Approver session.

| # | Request shape | Expected | Actual | Result |
|---|---|---|---|---|
| T1 | `content_hash=000…` + `If-Match:"v1"` | 412 `precondition.content_hash_mismatch` | `412 {"code":"precondition.content_hash_mismatch","title":"approval: content hash mismatch"}` | ✓ |
| T2 | `content_hash:""` + `If-Match:"v1"` | 412 / 400 | `400 {"code":"validation.request_invalid","title":"…content_hash is required"}` | ✓ (HTTP boundary; reaches before tier-2 check) |
| T3 | omitted `content_hash` + `If-Match:"v1"` | 412 / 400 | `400 {"code":"validation.request_invalid","title":"…content_hash is required"}` | ✓ (HTTP boundary) |
| T4 | correct hash + `If-Match:"v999"` | precondition fail (distinct from content-hash) | `409 {"code":"conflict.stale_revision","title":"…stale revision"}` | ✓ semantically; **note** 409 not 412 (briefing wording) |
| T5 | correct hash + `If-Match:1` (no quotes/no `v`) | 400 malformed | `400 {"code":"validation.if_match_malformed"}` | ✓ |
| T6 | correct hash, no `If-Match` | 400/428 | `400 {"code":"validation.header_required","title":"…If-Match is required"}` | ✓ |
| T7 | correct hash, no `Idempotency-Key` | 400 | `400 {"code":"validation.header_required","title":"…Idempotency-Key is required"}` | ✓ |
| T8 | `content_hash:"abc"` (3-char) | 400 64-hex | `400 {"code":"validation.request_invalid","title":"…content_hash must be 64 hex characters"}` | ✓ |
| T9 | wrong hash + `If-Match:*` wildcard (post-terminal) | precondition or state | `409 {"code":"state.instance_completed"}` (run after T11 happy path completed instance) | ✓ |

### 2.3 Happy path + idempotency replay (approver)

| # | Request | Expected | Actual | Result |
|---|---|---|---|---|
| T10 | `POST /documents/c1bb2112-…/signoff` `If-Match:"v1"` + correct `content_hash` + `Idempotency-Key:qa-happy-<X>` | 200 outcome=approved | `200 {"signoff_id":"","was_replay":false,"outcome":"approved"}` (107 ms) | ✓ |
| T11 | Same body + same idempotency key | 200 `was_replay:true, outcome:"approved"` | `409 {"code":"state.instance_completed"}` | **✗** see F-002 |
| T12 | Read instance after approval | `status=approved`, `etag="v2"` | `200 status=approved completed_at=2026-06-01T14:59:46Z etag="v2"` + signoff row populated | ✓ |

### 2.4 Negative post-terminal

| # | Request | Expected | Actual | Result |
|---|---|---|---|---|
| T13 | Admin signoff on already-approved instance | 409 instance terminal | `409 {"code":"state.instance_completed"}` | ✓ |
| T14 | Approver signoff on already-approved instance | 409 instance terminal | `409 {"code":"state.instance_completed"}` | ✓ |
| T15 | Cancel completed instance (admin, `If-Match:"v2"`) | 409 invalid transition | `409 {"code":"state.instance_completed"}` | ✓ |
| T16 | Resubmit approved doc (admin) | 4xx | `400 {"code":"validation.request_invalid","title":"…route_id is required"}` — validation blocks before state check | ✓ partial — body shape blocks, behavior on shaped resubmit not exercised |

### 2.5 Taxonomy authz (regression check for `8e1518c85`)

Approver session — has `document.view`, **no** `taxonomy.manage`.

| # | Request | Expected (per commit msg) | Actual | Result |
|---|---|---|---|---|
| TX1 | `GET /api/v1/taxonomy/families` | 200 | `403 {"code":"AUTH_FORBIDDEN","title":"Insufficient permissions"}` | **✗ CRITICAL F-001** |
| TX2 | `GET /api/v1/taxonomy/profiles` | 200 | `403 AUTH_FORBIDDEN` | **✗ F-001** |
| TX3 | `GET /api/v1/taxonomy/areas` | 200 | `403 AUTH_FORBIDDEN` | **✗ F-001** |
| TX4 | `GET /api/v1/taxonomy/family/po` | 200 | `404 page not found` — route does not exist (briefing path wrong; canonical is `/families/{code}`) | INFO route-shape clarification |
| TX5 | `GET /api/v1/taxonomy/profile/po` | 200 | `404 page not found` (same) | INFO |
| TX6 | `GET /api/v1/taxonomy/area/rh` | 200 | `404 page not found` (same) | INFO |
| TX7 | `GET /api/v1/taxonomy/areas/rh/ancestors` | 200 | `403 AUTH_FORBIDDEN` | **✗ F-001** |
| TX8 | `POST /api/v1/taxonomy/families` (write) | 403 still | `403 AUTH_FORBIDDEN` | ✓ writes still gated |

### 2.6 Persisted state (psql)

```
SELECT id,status,completed_at,submitted_by FROM approval_instances WHERE id='55e699f7-…';
→ id … | status=approved | completed_at=2026-06-01 14:59:46.185079+00 | submitted_by=admin
```

Governance events written same-tx with signoff:
```
event_type        | actor_user_id | created_at
signoff_recorded  | approver      | 2026-06-01 14:59:46.185079+00
authz.bypass_used | system:scheduler | …   (pre-existing system events)
```

Schema check (relates to commit `d4e7273da` mig 0216):
- `governance_events` columns: `id, tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json, created_at, dedupe_key, correlation_id` — schema present, INSERT succeeded ✓
- `approval_stage_instances.skip_reason` column **present** (`text NULL`) ✓
- Note: wiki/modules/approval.md claims governance event has `occurred_at`; actual column is `created_at`. Doc drift, not a code defect.

## 3. Findings

| ID | Sev | Area | Description | Evidence | Suggested fix |
|---|---|---|---|---|---|
| F-001 | **CRITICAL** | authz / taxonomy | Commit `8e1518c85` downgraded only the repository-tier (tier-2) authz in `internal/modules/taxonomy/infrastructure/{family_repository,repository}.go`. The HTTP-tier (tier-1) middleware permission registry at `apps/api/cmd/metaldocs-api/permissions.go:159-175` still gates **all** `/api/v1/taxonomy/{families,profiles,areas}` GETs on `CapTaxonomyManage`. Viewers (`document.view` only) still get `403 AUTH_FORBIDDEN` at the middleware before the repo ever runs. Stated intent of the commit is **not** delivered to viewers. | TX1/TX2/TX3/TX7 transcript above; `apps/api/cmd/metaldocs-api/permissions.go:159,165,171` all `MethodGet` rows reference `iamdomain.CapTaxonomyManage`. `permissions_test.go:90-94` asserts the same — tests are green because they encode the bug. | Add new GET-only permission entries with `CapDocumentView` (or whatever the canonical view-tier cap is) ahead of the existing prefix entries OR split the prefix into method-scoped entries (GET→View, POST/PATCH/PUT/DELETE→Manage). Update `permissions_test.go`. Verify with the same curl matrix above. |
| F-002 | MEDIUM | idempotency / signoff | After a doc-scoped signoff `POST /documents/{id}/signoff` terminates the instance (final approval), retrying with the **same Idempotency-Key + same body** returns `409 state.instance_completed` instead of the cached replay `{was_replay:true,outcome:"approved"}`. `SignoffByDocumentHandler` loads the active instance and shorts to `ErrInstanceCompleted` (handler `doc_approval_handler.go:124-128`) *before* calling `idempStore.BeginDocumentReplay` (line 139). Result: callers cannot safely retry a 200 they didn't observe. Network-loss → next request fails 409, looks like a real conflict. | T10/T11 transcript above; `internal/modules/documents/approval/http/doc_approval_handler.go:114-150` ordering. Stage-scoped surface `/approval/instances/{id}/stages/{sid}/signoffs` may have the same issue; not exercised this run. | Move `BeginDocumentReplay` ahead of the activeStage nil-check. The replay key already includes `(tenantID, actorID, idempKey, payloadHash)` and the comment at `handler.go:63` says the route template is part of the key — so a replay hit can safely return the cached outcome without re-reading instance state. Add a regression test for the post-terminal replay path. |
| F-003 | LOW | wiki drift | `wiki/modules/approval.md` "Same-tx governance events" bullet says `governance_events INSERT … explicit occurred_at` and `application/events.go:35` is cited — actual column is `created_at` (default `now()`). | `\d governance_events` output above. | Update the bullet and the `events.go` line citation. |
| F-004 | LOW | wiki drift | `wiki/references/local-dev-credentials.md` lists `approver-test` password `ApproverMetalDocs456!@`. Login returns 401 `AUTH_INVALID_CREDENTIALS`. Either password drifted or the seed migration `0166_*` (referenced as the rename event) wasn't re-applied locally. | Server log `2026/06/01 11:57:43 auth login failed for "approver-test"`. | Verify against `db/dev-seeds/0001_local_dev_seed.sql`; either bump the stamp + correct password OR re-run dev-seed and refresh the doc. |
| F-005 | INFO | content-pin layering | The `_content_hash` mandatory + fail-closed code path in `decision_service.go:174-185` is **unreachable** from the doc-scoped HTTP surface because `contracts.SignoffRequest.Validate()` rejects empty/missing `content_hash` first with `400 validation.request_invalid`. Defense-in-depth still intact (different layer), but the briefing's claim "Missing `_content_hash` field → expect 412" is **400 in practice**. Briefing wording vs. actual behavior is the discrepancy, not a code bug. | T2/T3 transcript above. | Update T-013 entry in `wiki/modules/approval-tech-debt.md` (or its closure note) to record that this layer catches first; only direct application-layer callers can hit the tier-2 check. |
| F-006 | INFO | OCC status code | Stale `If-Match` returns `409 conflict.stale_revision` (briefing said 412). Per RFC 7232 `If-Match` failure is canonically 412 `Precondition Failed`. Current 409 maps to `ErrStaleRevision` in `http/errors.go:82-84` regardless of header source. Not a regression — this is the documented internal mapping (`approvalCodeConflictStaleRevision`). Worth a one-liner in wiki/quality/backend-api-qa-checklist OR the approval module so future QA briefings match. | T4 transcript; `errors.go:82-84`. | Either change the mapping to 412 for `If-Match`-sourced staleness specifically, OR document the 409 convention in the module wiki. Doc-level fix is lower risk. |
| F-007 | INFO | non-tested branches | Stage-scoped signoff (`POST /api/v1/approval/instances/{id}/stages/{stage_id}/signoffs`) was not exercised this run — only the doc-scoped alias was, because the briefing's path `/api/v1/approvals/instances/{id}/signoff` (plural, no stage segment) is **not** a registered route (router.go:18-33). No regression, but the briefing should be corrected and the stage-scoped path covered in a follow-up. | `internal/modules/documents/approval/http/router.go:18-33`. | Add stage-scoped signoff matrix to the next QA brief. |

## 4. Static check outcomes

Run inside the prep workflow at session start (commit `8e1518c8`).

| Command | Exit | Outcome |
|---|---|---|
| `go build ./...` | 0 | pass — clean |
| `go build ./internal/modules/documents/approval/...` | 0 | pass — clean |
| `go test ./internal/modules/documents/approval/... -count=1 -timeout 120s` | 0 | pass — 8/8 packages (`application 3.318s, domain 3.197s, http 3.056s, http/contracts 2.879s, infrastructure 2.714s, infrastructure/signature 2.763s, jobs 2.858s, repository 2.857s`; `api` no test files) |
| `go test ./internal/modules/taxonomy/... -count=1 -timeout 120s` | 0 | pass — 5/5 packages |
| `go test -race` | DEFER | not run; gcc/race on Windows + time budget — see Defers |

## 5. Recent-commit regression confirmation

| Commit | Claim | Verdict |
|---|---|---|
| `392dd2051` align signoff content-pin source with active-document | content hash returned by `/active-document` is the **same** value accepted by `/documents/{id}/signoff`; defense-in-depth fails closed on wrong/missing hash | **✓** T1/T2/T3/T10 above; client-pinned hash matches server, happy path 200, wrong-hash 412 `precondition.content_hash_mismatch` |
| `d4e7273da` repair governance_events + stage skip_reason schema drift (mig 0216) | columns present, INSERTs succeed in tx with signoff | **✓** governance INSERT observed same-tx (`signoff_recorded` row appeared at `completed_at`); `approval_stage_instances.skip_reason` column present |
| `c9142a41b` log 5xx errors in WriteError | 5xx errors logged via `slog.Error` instead of swallowed | **✓ (code inspection)** `internal/modules/documents/approval/http/errors.go:230-236` does `slog.Error("approval handler error", slog.Int("status",…), slog.String("code",…), slog.Any("error", err))` when `prob.Status >= 500`. **Empirical** trigger of a 5xx not attempted (would need to break a dependency); no 5xx observed in this session's log tail (lines 17-64 of API stdout). |
| `8e1518c85` downgrade taxonomy read caps to View | viewers can GET taxonomy reads, writes still 403 | **✗ INEFFECTIVE — see F-001.** Repo-layer cap was downgraded but HTTP-tier-1 middleware still requires `CapTaxonomyManage`, so viewer requests are killed at the middleware before the repo runs. Writes correctly still 403 (correct outcome via wrong gate). |

## 6. Bounded defers

| Item | Reason | Follow-up |
|---|---|---|
| `go test -race ./...` | Windows local lacks gcc for race detector; not run | Re-run in CI / Linux runner |
| Fresh-doc `Submit` happy path | No draft docs in dev DB; creating a fresh draft requires template + CD + revision pipeline beyond QA scope. Submit semantics verified **indirectly** via pre-existing under_review instance (submitted_at 2026-06-01T01:48:38Z, stage 1 active, eligible actors snapshot present, etag `v1`) | A separate QA run should drive Submit → Inbox → Signoff end-to-end from a freshly created doc |
| Stage-scoped signoff endpoint (`/approval/instances/{id}/stages/{sid}/signoffs`) | Only one active instance; consumed by doc-scoped path. Briefing path was wrong (`/approvals/instances/{id}/signoff` does not exist) — see F-007 | Cover in next run with a fresh instance |
| 5xx logging empirical trigger | Cannot break a dependency in a shared dev env without collateral. Code-inspection only (`errors.go:230-236`) | Acceptable — code is straightforward and adjacent unit tests exist (`http/errors_test.go`) |
| `approver-test` login | Password drift (`401 AUTH_INVALID_CREDENTIALS`) — see F-004 | All approver-side tests done as `approver` instead; functionally equivalent (same caps + eligibility) |
| Multi-stage progression | Only route `QA Inbox Route` exists with 1 stage (`any_1_of`, quorum 0/1). Cannot validate stage-to-stage transition | Create a 2-stage route + fresh doc in a follow-up |
| `Publish` step after final approval | Doc moved to `status=approved` after signoff (verified). The publish/effective-date path was not separately exercised; PDF pdfOutbox enqueue path is async via `metaldocs-jobs` | Cover in next workflow-async QA run |

## 7. Verdict

**PASS-WITH-WARNINGS.**

Evidence:
- Signoff content-pin (`392dd2051`): live happy path + 4 negative content-hash variants + correct-hash 200 ✓
- Governance + skip_reason schema (`d4e7273da`): `\d governance_events` + `\d approval_stage_instances` + observed `signoff_recorded` INSERT ✓
- 5xx logging (`c9142a41b`): code present at `errors.go:230-236` ✓ (empirical trigger deferred — see §6)
- Taxonomy cap downgrade (`8e1518c85`): **regression CRITICAL** — middleware still requires `CapTaxonomyManage` for GETs — viewers blocked at tier-1. Stated intent of the commit is not delivered. See F-001.
- Approval lifecycle: inbox → active-document → signoff → instance approved → governance event written same-tx → OCC + idempotency-key enforcement + post-terminal guard all behaving correctly except idempotency-replay-after-terminal (F-002, MEDIUM).

Close-out rule honoured: all transcripts captured with command + response status + response body excerpt; DB state verified via `docker exec metaldocs-postgres psql`; static checks ran clean; bounded defers explicitly listed with reason.

Next-action priority for fix session:
1. F-001 — fix tier-1 middleware permissions for taxonomy GETs (CRITICAL — viewers currently locked out of document-browsing taxonomy).
2. F-002 — reorder `SignoffByDocumentHandler` so idempotency replay is consulted before active-stage state check (MEDIUM — silent breakage of safe-retry contract).
3. F-003 / F-004 — wiki drift cleanups (LOW).
4. F-005 / F-006 / F-007 — clarify briefings + add stage-scoped coverage (INFO).

---

## 8. Browser QA (preview, Chromium, `:4173`)

Driver: `preview_*` MCP tools. Real-user click/fill flow. Server-side state mutated through the FE, not synthetic curl. Adds end-user evidence on top of §2 API transcript.

### 8.1 Environment

| Item | Value |
|---|---|
| Preview server | `metaldocs-web` on `:4173` (Vite dev) |
| Date | 2026-06-01 |
| Test doc | `PO-RH-005` — `QA Browser Flow 2026-06-01` (created via UI wizard) |
| Document UUID | `9e6f4eb3-b3ed-44b0-9373-882ba32e30d7` |
| Controlled-doc UUID | `5872de2e-b644-41e2-846e-9a4901aee97b` |
| Approver caps observed (`/auth/me`) | `document.create,document.edit,document.signoff,document.submit,document.view,template.approve,template.review,template.view` — **no `taxonomy.manage`** |

### 8.2 Flow exercised end-to-end through the FE

| # | UI action | Network observed | Result |
|---|---|---|---|
| B1 | Admin login form → submit | `POST /api/v1/auth/login → 200` | session set ✓ |
| B2 | Documentos → "Novo documento" → wizard 4 steps (perfil=`po`, área=`rh`, template=Em branco, title=`QA Browser Flow 2026-06-01`, confirm checkbox, Criar) | `POST /api/v1/controlled-documents → 201` + editor load | code `PO-RH-005` allocated atomically, draft created, editor opens ✓ |
| B3 | Editor opens, "Submeter para revisão" click | `POST /api/v1/documents/{doc}/finalize → 201` + `GET /approval-instance → 200` | status flips to `EM REVISÃO`, "Próximos aprovadores: Approver Dev" rendered ✓ |
| B4 | Logout admin → login approver (form, password `ApproverMetalDocs123!`) | `POST /auth/logout → 204`, `POST /auth/login → 200`, `GET /auth/me → 200` | session as approver ✓ |
| B5 | Aprovações → inbox shows the new doc, "SUA FILA: 1 decisões pendentes" | `GET /api/v1/approval/inbox?limit=6 → 200` | inbox renders correctly with the doc ✓ |
| B6 | Click "Aprovar e assinar →" → SignoffDialog opens with Decisão (Aprovado), Motivo, Senha | dialog mounted client-side | OK ✓ |
| B7 | Fill `Motivo=Browser QA approve`, `Senha=ApproverMetalDocs123!` → click "Confirmar assinatura" | `GET /api/v1/controlled-documents/{cd}/active-document → 200` (FE fetches content_hash) → `POST /api/v1/documents/{doc}/signoff → 200 {"outcome":"approved","was_replay":false,"signoff_id":""}` | signoff persisted, inbox count drops to 0, screen renders "Nenhuma aprovação pendente." ✓ |

**Commit-level corroborations from browser run:**
- `392dd2051` content-pin: FE always fetches `/active-document` immediately before `POST /signoff` — same COALESCE source the server uses. Request order observed: `12900.779` `GET …/active-document → 200`, then `12900.780` `POST …/signoff → 200`. Pin source aligned client- and server-side ✓.
- `d4e7273da` governance schema: implicit — signoff returned 200 (would have 500'd inside the same-tx INSERT if the column/event-type drift were still present).
- `c9142a41b` 5xx logging: no 5xx triggered in this UI flow, so unchanged from §6 defer status.

### 8.3 F-001 — browser confirmation

Same browser session as approver, hitting taxonomy GETs from page context (carries `metaldocs_session` cookie):

```js
await fetch('/api/v1/taxonomy/profiles', {credentials:'include'}).then(r=>r.status)  // 403
await fetch('/api/v1/taxonomy/areas',    {credentials:'include'}).then(r=>r.status)  // 403
await fetch('/api/v1/taxonomy/families', {credentials:'include'}).then(r=>r.status)  // 403
```

Body (sample):
```json
{ "code": "AUTH_FORBIDDEN", "status": 403, "title": "Insufficient permissions" }
```

`GET /auth/me` for the same session returned `document.view` as one of the caps and no `taxonomy.manage`. Per commit `8e1518c85`'s stated intent ("downgrade read-path authz caps from Manage to View"), these three GETs should return 200. They do not. **F-001 confirmed at the real user surface, not just via curl.**

Note: the same UI flow (B2 above) reached `/api/v1/taxonomy/profiles → 200` and `/api/v1/taxonomy/areas → 200` while logged in as admin — admin holds `taxonomy.manage`, so middleware allows it. The regression is invisible to anyone testing under an admin/superuser; it only surfaces under viewer-class roles.

### 8.4 What the browser pass did NOT cover

- **F-002 idempotency replay-after-terminal**: not exercised through UI. The FE generates a fresh `Idempotency-Key` per dialog open and clears state on close, so a real retry-after-terminal needs synthetic conditions. Stays a code-inspection finding per §3.
- **Reject path** (`Devolver`): not exercised — only Approve flow tested.
- **Stage-scoped signoff endpoint**: FE always uses `/documents/{id}/signoff` (doc-scoped). Stage-scoped path remains untested under any driver.
- **`approver-test` user**: still 401, password drift unresolved (F-004).

### 8.5 Console / unrelated runtime noise

`preview_console_logs` and `preview_network` show:
- One `POST /auth/logout → 204 [FAILED: net::ERR_ABORTED]` per logout — harmless: client navigates away before the response is consumed; server-side cookie is cleared.
- A handful of `https://fonts.googleapis.com/css2?family=Calibri%20Light…` blocked by ORB — cosmetic, comes from the docx editor's font fallback; does not affect approval flow.
- Initial unauthenticated `GET /auth/me → 401` and pre-login `/feature-flags → 500` — expected pre-session.

None of these affect the approval-flow verdict.

### 8.6 Browser verdict

Browser pass does not invalidate the prior §1-§7 conclusions:
- **F-001 escalation reinforced:** now reproducible end-to-end through a real session cookie, not only via curl with manual Origin header. Severity stays **CRITICAL**.
- **Happy-path signoff** through the full FE chain (wizard → editor → submit → inbox → signoff dialog → confirm) **passes**, including the content-pin defense-in-depth path (FE fetches `/active-document` before sending the pin).
- All other commit-level checks remain as in §7.

Final verdict unchanged: **PASS-WITH-WARNINGS**, with F-001 still the only CRITICAL.
