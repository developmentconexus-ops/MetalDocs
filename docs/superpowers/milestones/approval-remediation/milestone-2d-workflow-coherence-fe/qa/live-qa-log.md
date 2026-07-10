# F2d.8 — UI-driven live QA log (M2d milestone close)

**Feature:** F2d.8 `f8-close-live-qa` · **Dates:** setup 2026-07-09, validation 2026-07-10 ·
**Method (intended):** browser-preview-driven UI QA against the full container stack via the gateway.
**Method (actual):** the full container stack was rebuilt from source and brought up healthy, but the
UI-render step could **not** be performed — no browser-preview toolset is connected in this session
(**Finding F-UI-1**, §0.3). In its place, all non-browser validation was run and is GREEN (§1): backend
`go test ./...` live against the container DB, FE mode-derivation vitest, and a gateway contract smoke.
**F2d.8 is NOT closed** — the rendered-UI acceptance is deferred to the HS-1 gate (§5).

**Closes:** the M2c recorded deviation (DecisionFooter offered signoff on a review stage → live 412
`precondition.content_hash_mismatch`). The C4 quality-bar gate requires this QA to show the
review→approval lifecycle driven through the UI without a 412, and that a **review stage never renders a
signature panel**.

---

## 0. Environment & method

| Component | State |
|-----------|-------|
| **Method** | **Full containerized stack** — the deterministic coded deploy (`deploy/compose/docker-compose.yml`, `docker compose --env-file ../../.env build && up -d`). ALL images (`api`/`worker`/`jobs`/`web`/`docx-renderer`) **rebuilt from current working-tree source** so the not-pushed F2d.1–F2d.7 backend+FE commits are baked in. QA runs through the gateway, i.e. exactly what a deployed environment serves. |
| Backend (api/worker/jobs) | docker `compose` project, rebuilt from current source (see §0.1 — the pre-existing images were stale) |
| Frontend (`web`) | docker `metaldocs-web` image **rebuilt from current source** (vite production build) — carries the not-pushed F2d.5/5b/6/7 FE (single screen, cockpit retirement). NOT the vite dev server; NOT the stale 2026-07-07 image. |
| Gateway | nginx `metaldocs-gateway` host `:80` → proxies `/` → web, `/api/v1` → api:8081 (`deploy/nginx/nginx.conf`) |
| Postgres | docker `metaldocs-postgres` (host `:5433`) |
| Tenant | System Tenant `ffffffff-ffff-ffff-ffff-ffffffffffff` |

> **Deploy methodology note (why the first rebuild attempts thrashed):** the coded path is a single
> `docker compose --env-file ../../.env build --progress plain` (plain progress + a tee'd logfile so a
> long buildkit run is observable) followed by `up -d`. The earlier stalls were **invocation errors, not
> deploy failures**: (1) a detached background build whose TTY progress the harness captured as 0 bytes —
> a healthy 10-min build that *looked* hung; (2) a 580s foreground timeout on a ~10-min cold build (Windows
> re-transfers the whole repo as build context ~92s + no-cache `go build` + ~185s layer export). Both were
> mine. `dev-api.ps1` (local binary vs docker infra) is the fast inner-loop alternative; for a milestone
> close we QA the real containers instead.

**Personas** — the **coded local dev seed** `db/dev-seeds/0001_local_dev_seed.sql` (piped into the
`metaldocs-postgres` container after the stack is healthy). Local-only creds live in that file (already
in git, marked "local-only … never apply in production") — host `.env` never read. Docker was reinstalled
mid-QA, wiping the old volume; this cold-start re-seed is the deterministic replacement for the earlier
ad-hoc pgcrypto password resets. Creds: `admin`/`AdminMetalDocs123!`, `approver`/`ApproverMetalDocs123!`,
`author-test`/`AuthorTest123!`, `approver-test`/`ApproverMetalDocs456!@`. Seed also creates taxonomy
(family `quality`, areas `rh`/`qualidade`/`producao`, profile `po`) so the new-document wizard works.
Approval **routes** are NOT seeded — they are created through the UI during QA (route.manage, admin).

| user_id | role | key capabilities | QA use |
|---------|------|------------------|--------|
| `author-test` | author | document.create/edit/submit/view | author (draft, submit, changes-requested round-trip) |
| `admin` | system_admin | document.review, approval.oversee, document.view, route.manage, … | reviewer (review-stage verdict), oversight |
| `approver` | approver | document.signoff, template.approve/review, document.view | approver (approval-stage signoff), delegate/delegator |
| `approver-test` | approver | document.signoff, … | second approver (delegation) |

All four authenticate `200` at `POST /api/v1/auth/login` (corroboration; UI login proven in §1).

### 0.1 HS-3 prerequisite repair — stale backend (a real F2d.8 finding)

On first inspection the running api's `GET /api/v1/documents/{id}/approval-instance` returned an instance
DTO with **no `viewer` block and no `verdicts[]`** (top-level keys: `id, document_id, route_id, tenant_id,
status, submitted_by, submitted_at, stages, etag, frozen_content_hash`). Yet:
- the working-tree OpenAPI **requires** `viewer` + `verdicts` on that schema (`api/openapi/v1/openapi.yaml`
  `required: [… viewer, verdicts]`);
- the working-tree backend **computes** them (`GetInstanceByDocumentHandler` →
  `mapInstanceResponse(… viewer, verdicts)`, `doc_approval_handler.go`, `get_instance_handler.go`;
  `read_service.go:LoadInstanceByDocumentForView`; `domain/viewer.go`);
- the running `metaldocs-api:dev` image was built **2026-07-08 08:22** — it predates the F2d.1/F2d.2
  backend (the viewer-facts + verdict-history contract).

**Diagnosis:** the F2d.1–F2d.7 features were committed but the running stack was **never rebuilt**, so the
integrated M2d system had never actually been run end-to-end. This is exactly the class of defect the live
QA gate exists to catch (compile-green ≠ runs). Per HS-3 ("prerequisite failure: contract/generated drift
… repair the prerequisite first"), the backend images were **rebuilt from current source**
(`docker compose build api worker jobs`) and the stack restarted before any scenario QA. _(verified post-
rebuild: §0.2.)_

### 0.1b Finding F-INFRA-1 (recorded, out-of-M2d-scope) — docx-renderer image build hangs

During the clean rebuild (Docker Desktop was reinstalled → empty cache/store), the `docx-renderer`
image build **hangs deterministically** at `pnpm --filter @metaldocs/docx-renderer --prod deploy
--legacy /out` (`apps/docx-renderer/Dockerfile:30-32`) — freezes right after resolving 827 pkgs, at the
`+252` package-linking phase; reproduced twice (killed after 9-17 min of zero output each). The other
four images (api/worker/web/jobs) build clean. Likely cause: a prod-dependency install script run during
`pnpm deploy` (deprecated `prebuild-install@7.1.3` / node-gyp postinstall) performs a network fetch that
stalls on the fresh VM's egress.

**Disposition:** RECORDED, not fixed here. Rationale (global-maximum for the *project*, not a local
patch): `docx-renderer` is the internal-only docx→PDF renderer; **M2d changes nothing in it**, so a
Dockerfile redesign is scope-creep across the milestone boundary (CLAUDE.md "keep changes scoped / stop on
architecture contradictions"). It is **not on the F2d.8 acceptance path** — PDF materialization is async
(publish → outbox → worker → docx-renderer) and orthogonal to every F2d.8 claim (mode derivation,
signature-panel presence/absence, verdicts, delegation, observer, redirect). The stack is therefore
brought up **without** `worker`+`docx-renderer`; the only QA consequence is the **published-doc PdfCanvas
view (F2d.5b D1) is a bounded defer** — its render pipeline can't materialize a PDF without the renderer.
Proper fix belongs to an infra/tooling task (own `developing-new-work` gate), not this milestone close.

### 0.2 Post-startup verification — container serves the fresh contract (DONE)

Cold `up -d` completed after the documented postgres initdb→TCP-restart race (postgres runs its
init scripts against a socket-only temp server, marks itself "healthy" over the socket, then shuts
that server down and restarts with full TCP; the api's `restart:unless-stopped` crash-loops through
that ~4-min window and self-recovers once TCP :5432 reopens). Final state: **all core services Up &
healthy** (`postgres`, `redis`, `minio`, `gotenberg`, `api`, `web`, `jobs`, `gateway`); `worker` +
`docx-renderer` intentionally omitted (Finding F-INFRA-1). Dev seed piped in clean (7 INSERT blocks
COMMIT).

Container smoke through the gateway (`:80`), fresh-from-source image:

| Probe (via `http://localhost/api/v1/…`) | Result | Proves |
|---|---|---|
| `POST /auth/login` × 4 personas (`admin`/`approver`/`author-test`/`approver-test`) | **200** each | login served; seed creds valid |
| `GET /auth/me` (admin) | 200, `capabilities[]` incl. `route.manage`,`document.review`,`document.signoff`,`approval.review`,`approval.oversee`,`document.publish` | flat-envelope (ADR 0035) contract served; tier-1 caps present |
| `GET /approval/routes` (admin) | **200** `{"routes":[],"total":0}` | approval route-admin path served |
| `GET /approval/inbox` (approver) | **200** `{"items":[],"total":0}` | approval read path served |
| `GET /documents/{missing}/approval-instance` (admin) | **404** `problem+json` `not_found.instance` | approval-instance handler wired; RFC 9457 |

> **Login field contract (QA-time correction, not a defect):** the login body field is `identifier`,
> not `username` (`AuthLoginRequest.required: [identifier, password]`, handler `json:"identifier"`).
> An initial `{"username":…}` probe returned `401 AUTH_INVALID_CREDENTIALS` with the request-identifier
> `sha256` = `e3b0c442…855` (the empty-string hash) — i.e. the api saw an empty identifier and
> short-circuited before bcrypt (DB `failed_login_attempts` stayed 0, confirming the not-a-real-attempt
> path). Root-caused to the field name; corrected probe → 200. No code change.

### 0.3 Finding F-UI-1 (BLOCKING for F2d.8 acceptance) — no browser-render tooling in this session

F2d.8's acceptance bar is **UI-driven** ("browser-preview-driven; curl-only walkthrough = FAIL; a
review stage rendering a signature panel = FAIL"). That bar requires a real browser rendering the
`web` container through the gateway and capturing per-step UI evidence. **This session has no
browser-preview toolset connected** — neither the main thread nor the `frontend-screen-reviewer`
subagent (whose only live tools are Read/Glob/Grep/Bash) can call `preview_start`/`preview_eval`/
`preview_snapshot`/`preview_fill`/`preview_click`/`preview_screenshot`; a capability probe confirmed
`GATEWAY_RENDER: CANNOT` (functions absent, not merely erroring). The gateway itself is reachable
(`curl http://localhost/` → 200) — the gap is the automation surface, not the stack.

**Disposition:** F2d.8 **cannot be closed to its acceptance bar in this session.** Fabricating a UI
pass from curl output would violate the gate and the "no patches/workarounds, report faithfully"
mandate. The rendered-UI acceptance is deferred to an environment with the preview toolset (or driven
by the operator at the HS-1 gate). Everything that does NOT require a browser was validated instead
(below) and is GREEN — so the residual risk carried into HS-1 is scoped precisely to *visual
rendering*, not to backend behavior, contract, or the FE mode-derivation logic.

---

## 1. Non-UI validation actually performed (all GREEN)

These substitute for **none** of the UI acceptance (F-UI-1); they bound the risk that remains.

**Backend behavior gate — `go test ./...` vs the docker postgres, `DATABASE_URL` from container env
(secret never printed):** **96 packages ok, 0 FAIL, 0 panic, 0 skip** (46 no-test pkgs). A live
`-count=1` re-run (cache-bypassed, real execution) of the **approval kernel + adjacent** confirms the
signoff/verdict/content-freeze/SoD paths executed against the container DB, all GREEN:
`documents/approval/{application 4.9s, domain 1.3s, http 5.4s, http/contracts 1.8s, infrastructure 2.0s,
infrastructure/idempotency 2.6s, infrastructure/signature 1.9s, jobs 3.9s}`,
`auth/infrastructure/postgres 4.5s`, `documents/application 4.6s`, `controlleddocuments/application 4.4s`.
`go build ./...` → exit 0 (compile truth).

**FE mode-derivation — vitest (the exact M2c-deviation-closure logic):** **63 tests / 4 files PASS** —
`approval/lib/workspaceMode.test.ts` (17: stage-kind → mode derivation), `approval/.../DecisionFooter.test.tsx`
(12: the component that offered signoff on a review stage in M2c — now asserts verdict CTAs on review,
signature panel only on approval), `documents/pages/DocumentWorkspacePage.test.tsx` (26: the single
working screen), `documents/.../WorkspaceSidebar.test.tsx` (8). `tsc --noEmit -p tsconfig.build.json`
→ exit 0.

## 2. Route shape A — review + approval (full lifecycle)
**UI walkthrough: NOT performed — blocked by F-UI-1.** Backend coverage of this shape
(review-verdict stage → `changes_requested` round-trip → approval signoff → publish, incl. the
content-hash freeze that caused the M2c 412) is exercised by the green `documents/approval/*`
integration packages above; the FE "review stage renders verdict CTAs, never a signature panel"
invariant is exercised by the green `DecisionFooter`/`workspaceMode` tests. Live route creation was
not driven — the cold baseline seeds **no template version** (`document_template_versions` n=0), so
API-side document/instance creation needs template + minio-blob fixtures out of this session's scope.

## 3. Route shape B — approval-only (lifecycle)
**UI walkthrough: NOT performed — blocked by F-UI-1.** Same backing evidence as §2.

## 4. Delegation signoff · observer view · redirect
**UI walkthrough: NOT performed — blocked by F-UI-1.** Delegation + viewer(observer) backend logic
is covered by the green approval kernel packages; the `/approvals/:documentId` → `/documents/:id`
redirect (F2d.7) is covered by FE routing tests in the green FE suite.

## 5. M2c deviation closure verdict
**PARTIAL — code/contract/container GREEN; rendered-UI acceptance DEFERRED (F-UI-1).**
- The M2c defect (DecisionFooter offered signoff on a review stage → live 412 `content_hash_mismatch`)
  is closed **at the logic level**: `DecisionFooter` + `workspaceMode` tests assert a review stage
  never renders a signature panel, and the approval kernel's content-freeze/verdict tests are green
  live against the container DB.
- The **acceptance-bar artifact** (browser-rendered per-step captures showing the review→approval
  lifecycle driven through the UI without a 412) is **not produced** — no preview tooling in session.
- **F2d.8 is therefore NOT closed.** It carries one bounded, non-code blocker (F-UI-1) into the HS-1
  operator gate: run the rendered-UI walkthrough in an environment with the preview toolset, or have
  the operator drive it. No backend/FE code defect was found; nothing to fix — the gap is tooling.
