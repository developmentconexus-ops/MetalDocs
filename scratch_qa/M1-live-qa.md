# M1 canonical-submit-backend — Live QA (docker stack, rebuilt binary)

Stack: `docker compose --env-file .env up -d --build api` (rebuilt with M1 code; postgres unchanged).
Actor: `admin` (system_admin), tenant `ffffffff-...`. Doc `48e00392-...` (profile `po`, area `rh`, revision_number 0, draft).

| # | Action | Request | Result | Proves |
|---|--------|---------|--------|--------|
| 1 | Zero-gov submit, stale binary | `POST /submit` If-Match `"v0"` body `{}` | 400 `validation.if_match_malformed` | old binary rejects v0 (pre-fix) |
| — | Rebuild api container | — | healthy | current code live |
| 2 | Zero-gov submit, no route seeded | `POST /submit` If-Match `"v0"` body `{}` | **409 `state.approval_route_missing`** | F1 resolution ran (profile→route), F3 clean sentinel (not 500); v0 accepted; empty body accepted (F2) |
| 3 | Seed active route for profile `po` | `POST /approval/routes` | 201 route `6d3e1372` | fixture |
| 4 | Zero-gov submit, route present | `POST /submit` If-Match `"v0"` body `{}` | **201 `{instance_id, etag:"v1"}`** | **core objective**: fresh draft rev0, zero client governance data → success |
| 5 | Doc state after submit | `GET /documents/{id}` | status `under_review`, revision_version `1`, revision_title `Criacao do documento` | draft→under_review CAS + v-bump + REV0 title default |
| 6 | Re-submit under_review doc | `POST /submit` If-Match `"v1"` body `{}` | **409 `conflict.duplicate_submission`** | idempotency/instance-unique backstop, clean sentinel (not 500) |

REV≥1 live path (revision-of-published) not driven live (long journey); covered by the F1 testdb integration suite (`submit_service_defaults_integration_test.go`).
No 500s observed. All domain sentinels map to RFC 9457 problem+json.
