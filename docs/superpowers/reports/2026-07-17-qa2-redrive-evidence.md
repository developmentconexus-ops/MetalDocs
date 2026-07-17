# QA-2 Re-drive Evidence — post QR-A/QR-B remediation (2026-07-17)

Hub-owned QA-2 per operator ruling "Opção 1": rebuild :80 stack on main HEAD
(f336f15a), execute the materialize/PDF replay runbook once, re-drive the
QA-1 FAIL journeys via API against the gateway. J2-PDF remains BLOCKED on
chip QR-C (pdf event contract defect found during this QA-2 — see §4).

Stack: 5 images rebuilt from f336f15a, all healthy via gateway :80.
Personas: dev-seed admin / author-test (wiki/references/local-dev-credentials.md).

## 1. Journey verdicts

| # | Journey (QA-1 defect) | QA-1 result | QA-2 result | Verdict |
|---|---|---|---|---|
| 1 | Blank-docx live 409 (C2 Option B) | untested live | GET /templates/6f326aea/versions/1/docx-url → **409** `UPLOAD_MISSING` "DOCX file not yet uploaded…" | **PASS** |
| 2 | J5 template route, no profile_code (F18) | 400 profile required | POST /approval/routes subject_kind=template → **201** route d9f32079, `"profile_code":null` | **PASS** |
| 3 | J5 negative: template route WITH profile_code | n/a | **400** "profile_code must be absent for template routes" | **PASS** |
| 4 | J5 negative: document route WITHOUT profile_code | n/a | **400** "profile_code is required" | **PASS** |
| 5 | F2 registry truth: unknown capability | route accepted drift caps | **400** "\"template.nonexistent_cap\" is not a registered capability" | **PASS** |
| 6 | RouteSummary sentinel kill (QR-A) | `""` sentinel risk | GET /approval/routes list: zero `"profile_code":""` occurrences; nullable serialization | **PASS** |
| 7 | J3 template submit (F22) | 500 "template version reader not configured" | POST submit-for-approval → **200** `{"instance_id":"1b005785…","version_status":"under_review"}` | **PASS** |
| 8 | J3 template signoff (F22) | unreachable (500 upstream) | POST signoff decision=approve → **200** `{"outcome":"instance_approved"}`; DB: version → `approved`, instance → `approved` | **PASS** |
| 9 | J5a document route positive | — | **409** `route.duplicate_profile` (QA-1 route bbbb2222 still active for tenant+po) — single-active-route invariant enforced correctly; positive-create already proven live by bbbb2222 itself | **PASS** (rule-correct) |
| 10 | J2-PDF end-to-end | dead-letter | **BLOCKED on QR-C** (F-QA2-1/F-QA2-2, chip task_159e3d9a dispatched) | pending |

## 2. Replay runbook execution (once, per QR-3 closure terms)

`scripts/replay-materialize-pdf-deadletters.sql` via docker exec psql: revived 2
dead-lettered materialize events; materialize leg GREEN attempt-1;
`tenants/{tid}/revisions/{docid}/frozen.docx` objects confirmed in MinIO.
Downstream pdf leg exposed F-QA2-1/F-QA2-2 (below) — worker env fix committed
f336f15a (compose worker Gotenberg+MinIO stanza, operator-ratified ops repair).

## 3. Ops repairs performed by hub (state, not product code)

1. `deploy/compose/docker-compose.yml` worker env stanza (commit f336f15a) —
   nil PDFConverter root cause; operator ruling "1." classified as deploy-config
   QA prerequisite.
2. Template 86c13e5e v1 status reset under_review→draft (in-tx GUC-asserted
   UPDATE) — **misdiagnosis, self-corrected**: QA-1's submit HAD created
   instance b6b05c67 keyed by template *version* id 70ad42d3 (two-level keying:
   ROUTE.subject_key=template_id, INSTANCE.subject_key=version_id), invisible
   to the earlier template-id query. The forced draft made state inconsistent;
   repaired in step 3.
3. Cancelled pre-remediation instance b6b05c67 (capability snapshot
   `document.signoff` on a template subject + selector role_in_fixed_area
   approver/rh = the exact pre-QR-A drift class) + skipped its stage +
   deactivated drift-era route aaaa1111. Clean re-drive then used a fresh
   post-QR-A route 5b5f0eed (template.approve, named_user admin).

DB tripwires behaved fail-closed throughout (raw UPDATE without
`metaldocs.asserted_caps` JSONB `[{"cap":…}]` → P0001) — defense-in-depth
live-confirmed as a side effect.

## 4. New findings (this QA-2)

| ID | Sev | Finding | Disposition |
|---|---|---|---|
| F-QA2-1 | ship-blocker | worker service.go silently marks pdf events published when pdfRunner nil | chip QR-C (task_159e3d9a) |
| F-QA2-2 | ship-blocker | buildPDFEvent omits final_docx_s3_key; staging table has no docx-key column → every live pdf event dead-letters "missing final_docx_s3_key" | chip QR-C |
| F-QA2-3 | minor | template submit maps `ux_approval_instances_active_subject` unique violation to 500 INTERNAL_ERROR "unknown database error" instead of 409 duplicate-active-submission; reachable via submit race window (status read → insert) | defer → ROADMAP |
| F-QA2-4 | minor | template signoff 200 body returns `"signoff_id":""` (empty) alongside outcome=instance_approved | defer → ROADMAP |
| F-QA2-5 | minor | POST /approval/routes 201 body sparse: only route_id+profile_code populated; name/version/active/stages/created_at zero-valued | defer → ROADMAP |

## 5. Verdict

QA-2 (API-drivable scope): **PASS** — QR-1/QR-2/QR-4 remediations live-verified;
blank-docx 409 live-verified. QA-2 overall **PARTIAL/OPEN** until QR-C merges and
hub re-drives J2-PDF fresh (2 existing dead-lettered pdf events f6f4712a/83db8465
intentionally left as-is; fresh journey post-merge).
