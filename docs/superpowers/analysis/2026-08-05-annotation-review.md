# 2026-08-05 — Row-by-row annotation review (Task 17b)

**Scope:** every operation in the generated `httpSurface` table
(`apps/api/cmd/metaldocs-api/httpsurface_gen.go`, produced from the OpenAPI
spec's `x-authz-capability` extensions) compared, one row at a time, against
what `apps/api/cmd/metaldocs-api/permissions.go`'s legacy `routeRules` table
resolves for the same method + concrete path, via `resolveRoutePermission`'s
actual ordered-match semantics (first rule to match — `pathExact`,
`pathPrefix`, `pathSuffix`, `contains`, `notSuffix` — wins).

This is Task 17b of the http-surface-protocol program (plan §11 step 6). It
is the sole mitigation for the program's one accepted residual risk: a
transcription error copied consistently into both the spec annotation and
its own conformance-suite expectation, which the conformance suite cannot
detect by construction (see plan §11 step 6 rationale, and the task brief).

## Method

1. Parsed the 147 operations directly out of `api/openapi/v1/openapi.yaml`
   (method, path, `operationId`, `x-authz-capability`, `tags`) — the same
   source `cmd/gen-http-surface` reads to produce `httpsurface_gen.go`. Row
   count matches the generated table's 147 entries exactly (`grep -c` on
   both), so left and right sides are counting the same operations.
2. Transcribed `routeRules` verbatim, in file order, into a small Python
   harness that reimplements `routeRule.matches` and
   `resolveRoutePermission`'s first-match-wins scan byte-for-byte (same
   field semantics: exact / prefix / suffix / contains / notSuffix, AND'd
   within a rule, first matching rule in table order wins). This removes
   the risk of me mis-tracing the ordered scan by eye across 147 rows.
3. For every operation, built one concrete request path by substituting
   each `{param}` segment in the OpenAPI path with a literal placeholder
   (path structure — prefixes/suffixes/contains — is unaffected by the
   actual parameter value), ran it through the harness, and diffed the
   resolved capability string against the spec's `x-authz-capability`.
4. `routeRules` was read as "the surviving record of decisions people made,"
   never as an oracle — per row, the harness records *which rule matched and
   why* (method/pathExact/pathPrefix/pathSuffix/contains/notSuffix), so an
   agreement is traceable to a specific rule, not just a boolean.

`routeRules` and `resolveRoutePermission` were not modified to build this
review — the harness is a read-only transcription living in
`.superpowers/sdd/_route_rules.py` (scratch, not part of this commit), used
only to drive the comparison recorded in the table below.

## Summary

- **147 / 147 operations compared.** Count of generated `httpSurface`
  entries (`grep -cE '^	"(GET|POST|PUT|PATCH|DELETE) '
  apps/api/cmd/metaldocs-api/httpsurface_gen.go`) = 147; count of `paths.*`
  operations in `api/openapi/v1/openapi.yaml` = 147. Matches — no count
  drift between spec and generated table.
- **AGREE: 147. DIFFERS: 0.** Every operation's `x-authz-capability`
  (or, for the 7 operations with no capability — the 4 public + 3
  session-required ones — its absence) matches exactly what
  `resolveRoutePermission` resolves for that method + path via `routeRules`'
  first-match-wins scan.
- **SUSPECTED TRANSCRIPTION ERROR rows: none.** No row required an
  "I believe routeRules over the spec" judgment call.
- **What this DOES establish:** for every one of the 147 operations, the
  capability the spec declares is the same capability the pre-existing,
  independently-authored `routeRules` table — built up over many prior
  tasks (F-001 read/write splits, F4 route-admin cross-tier fix, F-DELETE-SHAPE,
  F-QA4-1, BE-9, the M2b/M2c approval-runtime rows, etc., per the extensive
  inline commentary in `permissions.go`) — would also enforce. Two
  independently-produced classification schemes agree on all 147 rows.
- **What this does NOT establish:** that either classification is
  *correct* against product/security intent. Both `routeRules` and the
  `x-authz-capability` annotations could share the same wrong belief about
  what an operation should require (the residual-risk scenario this task
  exists to catch is exactly a shared error, not a disagreement) — a clean
  AGREE sweep does not disprove that scenario, it only shows no *divergence*
  between the two independently-built records was found. It also does not
  re-verify tier-2 (`authz.Require`) call sites, which is a separate
  concern the F4/BE-9/F-QA4-1 fixes referenced in `routeRules`' comments
  were about (those are cited as evidence `routeRules` itself was already
  reconciled against tier-2 in earlier work, not re-checked here).
- **`routeRules` was not treated as an oracle.** Per the brief, a DIFFERS
  would have needed a stated reason naming which side is believed, in one
  of three shapes (accidental routeRules match, deliberate Task 3a/3b
  change, or suspected transcription error). Zero DIFFERS were found to
  classify, so no such calls were needed here — see "does NOT establish"
  above for why that is not the same as clearing the risk entirely.
- **Visibility classes with no `routeRules` "row" in the normal sense —
  public / session-required operations (7 total: `checkLiveness`,
  `checkReadiness`, `getFeatureFlags` = public; `getCurrentUser`,
  `changePassword`, `login`, `logout` = session-required):** these ARE
  present in `routeRules` as explicit rows, just with `capability` left as
  the zero value (no capability gate) and only `visibility` set. They are
  included in the table below like every other row, with spec capability
  shown as "(none — visibility-only)" and compared on visibility instead
  of capability; all 7 AGREE (both sides declare the same visibility class
  and neither declares a capability).
- **One item found outside the 147-row scope, noted for the record only:**
  `routeRules` has a row — `GET /api/v1/signed` → `CapTemplateView`
  (permissions.go line 262, "Signed-URL relay") — with **no corresponding
  operation anywhere in the 147-entry generated table or the OpenAPI spec**.
  This is the opposite direction from what this task was scoped to check
  (spec → routeRules, not routeRules → spec), so it is not scored as a
  DIFFERS against any of the 147 rows above and nothing here was changed to
  address it. It is left as a flagged observation: either the route is
  dead code in `routeRules` (harmless — an extra unreachable rule) or a
  route exists in the mux without a spec entry (which would be a real gap,
  outside this task's boundary). Not fixed, per the brief's "write it down,
  don't fix it" instruction.

## Table

Columns: `operationId (METHOD path)` | spec capability | routeRules row
(rule that matched, per `resolveRoutePermission`'s first-match order) →
resolved capability | AGREE/DIFFERS | reason (only populated for DIFFERS).
Grouped by the spec's own tag ordering (`specTags` in
`httpsurface_gen.go`). Paths below omit the `/api/v1` prefix to match the
OpenAPI spec's own path keys; the `routeRules row` column shows the full
`/api/v1/...` path fragments as they appear in `permissions.go`, since that
is what the rule literally matches against.

### approval (20)

| `revokeApprovalDelegation` (DELETE /approval/delegations/{id}) | CapDocumentSignoff | DELETE prefix=/api/v1/approval/delegations/ -> CapDocumentSignoff | AGREE | |
| `listApprovalInbox` (GET /approval/inbox) | CapDocumentView | GET exact=/api/v1/approval/inbox -> CapDocumentView | AGREE | |
| `getApprovalInstance` (GET /approval/instances/{instance_id}) | CapDocumentView | GET prefix=/api/v1/approval/instances/ -> CapDocumentView | AGREE | |
| `listApprovalRoutes` (GET /approval/routes) | CapRouteManage | GET prefix=/api/v1/approval/routes -> CapRouteManage | AGREE | |
| `getApprovalInstanceByDocument` (GET /documents/{id}/approval-instance) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocumentApprovalPreview` (GET /documents/{id}/approval-preview) | CapDocumentSubmit | GET prefix=/api/v1/documents suffix=/approval-preview -> CapDocumentSubmit | AGREE | |
| `createApprovalDelegation` (POST /approval/delegations) | CapDocumentSignoff | POST exact=/api/v1/approval/delegations -> CapDocumentSignoff | AGREE | |
| `cancelApprovalInstance` (POST /approval/instances/{instance_id}/cancel) | CapDocumentEdit | POST prefix=/api/v1/approval/instances/ suffix=/cancel -> CapDocumentEdit | AGREE | |
| `extendApprovalSLA` (POST /approval/instances/{instance_id}/extend-sla) | CapApprovalSLAExtend | POST prefix=/api/v1/approval/instances/ suffix=/extend-sla -> CapApprovalSLAExtend | AGREE | |
| `recordApprovalFastForward` (POST /approval/instances/{instance_id}/stages/{stage_id}/fast-forward) | CapDocumentSignoff | POST prefix=/api/v1/approval/instances/ suffix=/fast-forward -> CapDocumentSignoff | AGREE | |
| `recordApprovalReviewVerdict` (POST /approval/instances/{instance_id}/stages/{stage_id}/review-verdict) | CapApprovalReview | POST prefix=/api/v1/approval/instances/ suffix=/review-verdict -> CapApprovalReview | AGREE | |
| `recordApprovalStageSignoff` (POST /approval/instances/{instance_id}/stages/{stage_id}/signoffs) | CapDocumentSignoff | POST prefix=/api/v1/approval/instances/ suffix=/signoffs -> CapDocumentSignoff | AGREE | |
| `createApprovalRoute` (POST /approval/routes) | CapRouteManage | POST prefix=/api/v1/approval/routes -> CapRouteManage | AGREE | |
| `deactivateApprovalRoute` (POST /approval/routes/{id}/deactivate) | CapRouteManage | POST prefix=/api/v1/approval/routes -> CapRouteManage | AGREE | |
| `cancelDocumentApproval` (POST /documents/{id}/cancel) | CapDocumentEdit | POST prefix=/api/v1/documents suffix=/cancel -> CapDocumentEdit | AGREE | |
| `obsoleteDocument` (POST /documents/{id}/obsolete) | CapDocumentObsolete | POST prefix=/api/v1/documents suffix=/obsolete -> CapDocumentObsolete | AGREE | |
| `markDocumentReviewed` (POST /documents/{id}/review) | CapDocumentReview | POST prefix=/api/v1/documents suffix=/review -> CapDocumentReview | AGREE | |
| `recordDocumentSignoff` (POST /documents/{id}/signoff) | CapDocumentSignoff | POST prefix=/api/v1/documents suffix=/signoff -> CapDocumentSignoff | AGREE | |
| `submitDocumentForApproval` (POST /documents/{id}/submit) | CapDocumentSubmit | POST prefix=/api/v1/documents suffix=/submit -> CapDocumentSubmit | AGREE | |
| `updateApprovalRoute` (PUT /approval/routes/{id}) | CapRouteManage | PUT prefix=/api/v1/approval/routes -> CapRouteManage | AGREE | |

### audit (4)

| `listAuditEvents` (GET /audit/events) | CapAuditRead | GET exact=/api/v1/audit/events -> CapAuditRead | AGREE | |
| `getAuditExportStatus` (GET /audit/events/export/{export_id}) | CapAuditRead | GET prefix=/api/v1/audit/events/export/ -> CapAuditRead | AGREE | |
| `downloadAuditExport` (GET /audit/events/export/{export_id}/download) | CapAuditRead | GET prefix=/api/v1/audit/events/export/ -> CapAuditRead | AGREE | |
| `exportAuditEvents` (POST /audit/events/export) | CapAuditRead | POST exact=/api/v1/audit/events/export -> CapAuditRead | AGREE | |

### auth (4)

| `getCurrentUser` (GET /auth/me) | (none — visibility-only) | GET exact=/api/v1/auth/me -> (none) | AGREE | |
| `changePassword` (POST /auth/change-password) | (none — visibility-only) | POST exact=/api/v1/auth/change-password -> (none) | AGREE | |
| `login` (POST /auth/login) | (none — visibility-only) | POST exact=/api/v1/auth/login -> (none) | AGREE | |
| `logout` (POST /auth/logout) | (none — visibility-only) | POST exact=/api/v1/auth/logout -> (none) | AGREE | |

### configuration (1)

| `getFeatureFlags` (GET /feature-flags) | (none — visibility-only) | GET exact=/api/v1/feature-flags -> (none) | AGREE | |

### controlled-documents (9)

| `listControlledDocuments` (GET /controlled-documents) | CapDocumentView | GET prefix=/api/v1/controlled-documents -> CapDocumentView | AGREE | |
| `getControlledDocumentCreationContext` (GET /controlled-documents/creation-context) | CapControlledDocumentCreate | GET exact=/api/v1/controlled-documents/creation-context -> CapControlledDocumentCreate | AGREE | |
| `previewControlledDocumentCode` (GET /controlled-documents/preview-code) | CapControlledDocumentCreate | GET exact=/api/v1/controlled-documents/preview-code -> CapControlledDocumentCreate | AGREE | |
| `getControlledDocument` (GET /controlled-documents/{id}) | CapDocumentView | GET prefix=/api/v1/controlled-documents -> CapDocumentView | AGREE | |
| `getActiveDocument` (GET /controlled-documents/{id}/active-document) | CapDocumentView | GET prefix=/api/v1/controlled-documents -> CapDocumentView | AGREE | |
| `atomicCreateControlledDocument` (POST /controlled-documents) | CapControlledDocumentCreate | POST exact=/api/v1/controlled-documents -> CapControlledDocumentCreate | AGREE | |
| `createControlledDocumentRevision` (POST /controlled-documents/{id}/revisions) | CapDocumentEdit | POST prefix=/api/v1/controlled-documents suffix=/revisions -> CapDocumentEdit | AGREE | |
| `obsoleteControlledDocument` (PUT /controlled-documents/{id}/obsolete) | CapControlledDocumentObsolete | PUT prefix=/api/v1/controlled-documents suffix=/obsolete -> CapControlledDocumentObsolete | AGREE | |
| `supersedeControlledDocument` (PUT /controlled-documents/{id}/supersede) | CapControlledDocumentSupersede | PUT prefix=/api/v1/controlled-documents suffix=/supersede -> CapControlledDocumentSupersede | AGREE | |

### distribution (3)

| `getDocumentDistribution` (GET /documents/{id}/distribution) | CapDistributionRead | GET prefix=/api/v1/documents suffix=/distribution -> CapDistributionRead | AGREE | |
| `getDocumentDistributionCoverage` (GET /documents/{id}/distribution/coverage) | CapDistributionRead | GET prefix=/api/v1/documents suffix=/distribution/coverage -> CapDistributionRead | AGREE | |
| `listDocumentDistributionRecipients` (GET /documents/{id}/distribution/recipients) | CapDistributionRead | GET prefix=/api/v1/documents suffix=/distribution/recipients -> CapDistributionRead | AGREE | |

### documents (29)

| `deleteDocumentComment` (DELETE /documents/{id}/comments/{library_id}) | CapDocumentEdit | DELETE prefix=/api/v1/documents contains=/comments/ -> CapDocumentEdit | AGREE | |
| `listDocuments` (GET /documents) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `documentStats` (GET /documents/stats) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocument` (GET /documents/{id}) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `listDocumentCheckpoints` (GET /documents/{id}/checkpoints) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `listDocumentComments` (GET /documents/{id}/comments) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocumentDocxURL` (GET /documents/{id}/export/docx-url) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocumentFillInSchema` (GET /documents/{id}/fill-in-schema) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocumentPlaceholderOptions` (GET /documents/{id}/placeholder-options/{pid}) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `listDocumentPlaceholderValues` (GET /documents/{id}/placeholders) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocumentRevisionHistory` (GET /documents/{id}/revision-history) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `getDocumentRevisionUrl` (GET /documents/{id}/revisions/{rid}/url) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `viewDocument` (GET /documents/{id}/view) | CapDocumentView | GET prefix=/api/v1/documents -> CapDocumentView | AGREE | |
| `renameDocument` (PATCH /documents/{id}) | CapDocumentEdit | PATCH prefix=/api/v1/documents -> CapDocumentEdit | AGREE | |
| `updateDocumentComment` (PATCH /documents/{id}/comments/{library_id}) | CapDocumentEdit | PATCH prefix=/api/v1/documents -> CapDocumentEdit | AGREE | |
| `archiveDocument` (POST /documents/{id}/archive) | CapDocumentEdit | POST prefix=/api/v1/documents suffix=/archive -> CapDocumentEdit | AGREE | |
| `commitDocumentAutosave` (POST /documents/{id}/autosave/commit) | CapDocumentEdit | POST prefix=/api/v1/documents contains=/autosave/ -> CapDocumentEdit | AGREE | |
| `presignDocumentAutosave` (POST /documents/{id}/autosave/presign) | CapDocumentEdit | POST prefix=/api/v1/documents contains=/autosave/ -> CapDocumentEdit | AGREE | |
| `createDocumentCheckpoint` (POST /documents/{id}/checkpoints) | CapDocumentEdit | POST prefix=/api/v1/documents contains=/checkpoints -> CapDocumentEdit | AGREE | |
| `restoreDocumentCheckpoint` (POST /documents/{id}/checkpoints/{version}/restore) | CapDocumentEdit | POST prefix=/api/v1/documents suffix=/restore contains=/checkpoints/ -> CapDocumentEdit | AGREE | |
| `createDocumentComment` (POST /documents/{id}/comments) | CapDocumentEdit | POST prefix=/api/v1/documents suffix=/comments -> CapDocumentEdit | AGREE | |
| `duplicateDocument` (POST /documents/{id}/duplicate) | CapDocumentCreate | POST prefix=/api/v1/documents suffix=/duplicate -> CapDocumentCreate | AGREE | |
| `exportDocumentPDF` (POST /documents/{id}/export/pdf) | CapDocumentView | POST prefix=/api/v1/documents suffix=/export/pdf -> CapDocumentView | AGREE | |
| `reconstructDocument` (POST /documents/{id}/reconstruct) | CapDocumentEdit | POST prefix=/api/v1/documents suffix=/reconstruct -> CapDocumentEdit | AGREE | |
| `acquireDocumentSession` (POST /documents/{id}/session/acquire) | CapDocumentEdit | POST prefix=/api/v1/documents contains=/session/ -> CapDocumentEdit | AGREE | |
| `forceReleaseDocumentSession` (POST /documents/{id}/session/force-release) | CapMembershipManage | POST prefix=/api/v1/documents contains=/session/force-release -> CapMembershipManage | AGREE | |
| `heartbeatDocumentSession` (POST /documents/{id}/session/heartbeat) | CapDocumentEdit | POST prefix=/api/v1/documents contains=/session/ -> CapDocumentEdit | AGREE | |
| `releaseDocumentSession` (POST /documents/{id}/session/release) | CapDocumentEdit | POST prefix=/api/v1/documents contains=/session/ -> CapDocumentEdit | AGREE | |
| `putDocumentPlaceholderValue` (PUT /documents/{id}/placeholders/{pid}) | CapDocumentEdit | PUT prefix=/api/v1/documents contains=/placeholders/ -> CapDocumentEdit | AGREE | |

### health (2)

| `checkLiveness` (GET /health/live) | (none — visibility-only) | GET exact=/api/v1/health/live -> (none) | AGREE | |
| `checkReadiness` (GET /health/ready) | (none — visibility-only) | GET exact=/api/v1/health/ready -> (none) | AGREE | |

### iam (26)

| `revokeSession` (DELETE /auth/sessions/{session_id}) | CapSessionManage | DELETE prefix=/api/v1/auth/sessions/ -> CapSessionManage | AGREE | |
| `revokeAreaMembership` (DELETE /iam/area-memberships/{user_id}/{area_code}) | CapMembershipManage | DELETE prefix=/api/v1/iam/area-memberships/ -> CapMembershipManage | AGREE | |
| `listSessions` (GET /auth/sessions) | CapUserView | GET exact=/api/v1/auth/sessions -> CapUserView | AGREE | |
| `getIamAdminOverview` (GET /iam/admin/overview) | CapUserView | GET exact=/api/v1/iam/admin/overview -> CapUserView | AGREE | |
| `listAreaMemberships` (GET /iam/area-memberships) | CapMembershipView | GET exact=/api/v1/iam/area-memberships -> CapMembershipView | AGREE | |
| `listCapabilities` (GET /iam/capabilities) | CapMembershipView | GET exact=/api/v1/iam/capabilities -> CapMembershipView | AGREE | |
| `getKpi` (GET /iam/kpi) | CapMetricsView | GET exact=/api/v1/iam/kpi -> CapMetricsView | AGREE | |
| `getPresenceSnapshot` (GET /iam/presence/snapshot) | CapUserView | GET exact=/api/v1/iam/presence/snapshot -> CapUserView | AGREE | |
| `streamPresence` (GET /iam/presence/stream) | CapUserView | GET exact=/api/v1/iam/presence/stream -> CapUserView | AGREE | |
| `listRoleCapabilities` (GET /iam/role-capabilities) | CapMembershipView | GET exact=/api/v1/iam/role-capabilities -> CapMembershipView | AGREE | |
| `listRoles` (GET /iam/roles) | CapMembershipView | GET exact=/api/v1/iam/roles -> CapMembershipView | AGREE | |
| `getUsage` (GET /iam/usage) | CapMetricsView | GET exact=/api/v1/iam/usage -> CapMetricsView | AGREE | |
| `listUsers` (GET /iam/users) | CapUserView | GET exact=/api/v1/iam/users -> CapUserView | AGREE | |
| `listMemberships` (GET /iam/users/{user_id}/memberships) | CapMembershipView | GET prefix=/api/v1/iam/users/ suffix=/memberships -> CapMembershipView | AGREE | |
| `patchUser` (PATCH /iam/users/{user_id}) | CapUserManage | PATCH prefix=/api/v1/iam/users/ notSuffix=/roles -> CapUserManage | AGREE | |
| `grantAreaMembership` (POST /iam/area-memberships) | CapMembershipManage | POST exact=/api/v1/iam/area-memberships -> CapMembershipManage | AGREE | |
| `createManagedUser` (POST /iam/users) | CapUserManage | POST exact=/api/v1/iam/users -> CapUserManage | AGREE | |
| `bulkUsers` (POST /iam/users/bulk) | CapUserManage | POST exact=/api/v1/iam/users/bulk -> CapUserManage | AGREE | |
| `inviteUser` (POST /iam/users/invite) | CapUserManage | POST exact=/api/v1/iam/users/invite -> CapUserManage | AGREE | |
| `resetPassword` (POST /iam/users/{user_id}/reset-password) | CapUserManage | POST prefix=/api/v1/iam/users/ suffix=/reset-password -> CapUserManage | AGREE | |
| `upsertUserRole` (POST /iam/users/{user_id}/roles) | CapUserManage | POST prefix=/api/v1/iam/users/ suffix=/roles -> CapUserManage | AGREE | |
| `unlockUser` (POST /iam/users/{user_id}/unlock) | CapUserManage | POST prefix=/api/v1/iam/users/ suffix=/unlock -> CapUserManage | AGREE | |
| `onboardTenant` (POST /tenants) | CapTenantOnboard | POST exact=/api/v1/tenants -> CapTenantOnboard | AGREE | |
| `eraseTenant` (POST /tenants/{tenant_id}/erase) | CapTenantErase | POST prefix=/api/v1/tenants/ suffix=/erase -> CapTenantErase | AGREE | |
| `exportTenant` (POST /tenants/{tenant_id}/export) | CapTenantExport | POST prefix=/api/v1/tenants/ suffix=/export -> CapTenantExport | AGREE | |
| `replaceUserRoles` (PUT /iam/users/{user_id}/roles) | CapUserManage | PUT prefix=/api/v1/iam/users/ suffix=/roles -> CapUserManage | AGREE | |

### notifications (4)

| `listNotifications` (GET /notifications) | CapNotificationRead | GET exact=/api/v1/notifications -> CapNotificationRead | AGREE | |
| `getNotificationsUnreadCount` (GET /notifications/unread-count) | CapNotificationRead | GET exact=/api/v1/notifications/unread-count -> CapNotificationRead | AGREE | |
| `markAllNotificationsRead` (POST /notifications/read-all) | CapNotificationRead | POST exact=/api/v1/notifications/read-all -> CapNotificationRead | AGREE | |
| `markNotificationRead` (POST /notifications/{id}/read) | CapNotificationRead | POST prefix=/api/v1/notifications/ suffix=/read -> CapNotificationRead | AGREE | |

### observability (1)

| `getMetrics` (GET /metrics) | CapMetricsView | GET exact=/api/v1/metrics -> CapMetricsView | AGREE | |

### search (1)

| `searchDocuments` (GET /search/documents) | CapDocumentView | GET exact=/api/v1/search/documents -> CapDocumentView | AGREE | |

### security (3)

| `listLockouts` (GET /security/lockouts) | CapUserView | GET exact=/api/v1/security/lockouts -> CapUserView | AGREE | |
| `getMfaCoverage` (GET /security/mfa-coverage) | CapUserView | GET exact=/api/v1/security/mfa-coverage -> CapUserView | AGREE | |
| `listSecuritySignals` (GET /security/signals) | CapUserView | GET exact=/api/v1/security/signals -> CapUserView | AGREE | |

### taxonomy (16)

| `archiveTaxonomyArea` (DELETE /taxonomy/areas/{code}) | CapTaxonomyManage | DELETE prefix=/api/v1/taxonomy/areas -> CapTaxonomyManage | AGREE | |
| `deactivateTaxonomyFamily` (DELETE /taxonomy/families/{code}) | CapTaxonomyManage | DELETE prefix=/api/v1/taxonomy/families -> CapTaxonomyManage | AGREE | |
| `archiveTaxonomyProfile` (DELETE /taxonomy/profiles/{code}) | CapTaxonomyManage | DELETE prefix=/api/v1/taxonomy/profiles -> CapTaxonomyManage | AGREE | |
| `listTaxonomyAreas` (GET /taxonomy/areas) | CapTaxonomyView | GET prefix=/api/v1/taxonomy/areas -> CapTaxonomyView | AGREE | |
| `getTaxonomyArea` (GET /taxonomy/areas/{code}) | CapTaxonomyView | GET prefix=/api/v1/taxonomy/areas -> CapTaxonomyView | AGREE | |
| `listTaxonomyFamilies` (GET /taxonomy/families) | CapTaxonomyView | GET prefix=/api/v1/taxonomy/families -> CapTaxonomyView | AGREE | |
| `getTaxonomyFamily` (GET /taxonomy/families/{code}) | CapTaxonomyView | GET prefix=/api/v1/taxonomy/families -> CapTaxonomyView | AGREE | |
| `listTaxonomyProfiles` (GET /taxonomy/profiles) | CapTaxonomyView | GET prefix=/api/v1/taxonomy/profiles -> CapTaxonomyView | AGREE | |
| `getTaxonomyProfile` (GET /taxonomy/profiles/{code}) | CapTaxonomyView | GET prefix=/api/v1/taxonomy/profiles -> CapTaxonomyView | AGREE | |
| `updateTaxonomyFamily` (PATCH /taxonomy/families/{code}) | CapTaxonomyManage | PATCH prefix=/api/v1/taxonomy/families -> CapTaxonomyManage | AGREE | |
| `updateTaxonomyProfile` (PATCH /taxonomy/profiles/{code}) | CapTaxonomyManage | PATCH prefix=/api/v1/taxonomy/profiles -> CapTaxonomyManage | AGREE | |
| `createTaxonomyArea` (POST /taxonomy/areas) | CapTaxonomyManage | POST prefix=/api/v1/taxonomy/areas -> CapTaxonomyManage | AGREE | |
| `createTaxonomyFamily` (POST /taxonomy/families) | CapTaxonomyManage | POST prefix=/api/v1/taxonomy/families -> CapTaxonomyManage | AGREE | |
| `createTaxonomyProfile` (POST /taxonomy/profiles) | CapTaxonomyManage | POST prefix=/api/v1/taxonomy/profiles -> CapTaxonomyManage | AGREE | |
| `updateTaxonomyArea` (PUT /taxonomy/areas/{code}) | CapTaxonomyManage | PUT prefix=/api/v1/taxonomy/areas -> CapTaxonomyManage | AGREE | |
| `setTaxonomyProfileDefaultTemplate` (PUT /taxonomy/profiles/{code}/default-template) | CapTaxonomyManage | PUT prefix=/api/v1/taxonomy/profiles -> CapTaxonomyManage | AGREE | |

### templates (19)

| `listTemplates` (GET /templates) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `listTemplatePlaceholderCatalog` (GET /templates/placeholder-catalog) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `getSystemBlankTemplate` (GET /templates/system/blank) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `getTemplate` (GET /templates/{id}) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `listTemplateAudit` (GET /templates/{id}/audit) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `getTemplateVersion` (GET /templates/{id}/versions/{n}) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `getTemplateVersionApprovalPreview` (GET /templates/{id}/versions/{n}/approval-preview) | CapTemplateSubmit | GET prefix=/api/v1/templates suffix=/approval-preview -> CapTemplateSubmit | AGREE | |
| `getTemplateDocxUrl` (GET /templates/{id}/versions/{n}/docx-url) | CapTemplateView | GET prefix=/api/v1/templates -> CapTemplateView | AGREE | |
| `createTemplate` (POST /templates) | CapTemplateCreate | POST exact=/api/v1/templates -> CapTemplateCreate | AGREE | |
| `archiveTemplate` (POST /templates/{id}/archive) | CapTemplateArchive | POST prefix=/api/v1/templates suffix=/archive -> CapTemplateArchive | AGREE | |
| `createTemplateVersion` (POST /templates/{id}/versions) | CapTemplateCreate | POST prefix=/api/v1/templates suffix=/versions -> CapTemplateCreate | AGREE | |
| `commitTemplateAutosave` (POST /templates/{id}/versions/{n}/autosave/commit) | CapTemplateEdit | POST prefix=/api/v1/templates contains=/autosave/ -> CapTemplateEdit | AGREE | |
| `presignTemplateAutosave` (POST /templates/{id}/versions/{n}/autosave/presign) | CapTemplateEdit | POST prefix=/api/v1/templates contains=/autosave/ -> CapTemplateEdit | AGREE | |
| `presignTemplateDocxUploadUrl` (POST /templates/{id}/versions/{n}/docx-upload-url) | CapTemplateEdit | POST prefix=/api/v1/templates suffix=/docx-upload-url -> CapTemplateEdit | AGREE | |
| `publishTemplateVersion` (POST /templates/{id}/versions/{n}/publish) | CapTemplatePublish | POST prefix=/api/v1/templates suffix=/publish -> CapTemplatePublish | AGREE | |
| `presignTemplateSchemaUploadUrl` (POST /templates/{id}/versions/{n}/schema-upload-url) | CapTemplateEdit | POST prefix=/api/v1/templates suffix=/schema-upload-url -> CapTemplateEdit | AGREE | |
| `signoffTemplateVersion` (POST /templates/{id}/versions/{n}/signoff) | CapTemplateApprove | POST prefix=/api/v1/templates suffix=/signoff -> CapTemplateApprove | AGREE | |
| `submitTemplateVersionForApproval` (POST /templates/{id}/versions/{n}/submit-for-approval) | CapTemplateSubmit | POST prefix=/api/v1/templates suffix=/submit-for-approval -> CapTemplateSubmit | AGREE | |
| `updateTemplateSchema` (PUT /templates/{id}/versions/{n}/schema) | CapTemplateEdit | PUT prefix=/api/v1/templates suffix=/schema -> CapTemplateEdit | AGREE | |

### tokens (5)

| `deleteToken` (DELETE /tokens/{id}) | CapTokenDictionaryManage | DELETE prefix=/api/v1/tokens -> CapTokenDictionaryManage | AGREE | |
| `listTokens` (GET /tokens) | CapTokenView | GET prefix=/api/v1/tokens -> CapTokenView | AGREE | |
| `getToken` (GET /tokens/{id}) | CapTokenView | GET prefix=/api/v1/tokens -> CapTokenView | AGREE | |
| `createToken` (POST /tokens) | CapTokenDictionaryManage | POST exact=/api/v1/tokens -> CapTokenDictionaryManage | AGREE | |
| `updateToken` (PUT /tokens/{id}) | CapTokenDictionaryManage | PUT prefix=/api/v1/tokens -> CapTokenDictionaryManage | AGREE | |
