# Observability + IAM Admin Panel — Deep Research

> **Audience:** MetalDocs design + frontend + backend engineers redesigning `/admin` (current screen: `frontend/apps/web/src/features/iam/AdminCenterView.tsx`).
> **Goal:** Replace the amateur three-KPI / online / activity / managed-users layout with a professional SaaS admin control panel grounded in industry references and bounded by MetalDocs ADR 0016 (view-grade capabilities) and the 8 canonical IAM roles.
> **Output type:** Research synthesis. Final section produces a concrete tab/section tree to implement against.
> **Research date:** 2026-06-02. All citations dated within the last 24 months unless otherwise noted.
> **Methodology:** Web search + doc fetch across 8 industry references; pattern synthesis; mapping to MetalDocs view-grade caps from ADR 0016; explicit anti-patterns harvested from the current screen.

---

## 0. Executive Summary

A professional SaaS admin/observability surface in 2026 is not a "users page". It is a **multi-tab workspace** that exposes — at minimum — six first-class surfaces:

1. **Overview** — calm, decision-oriented health snapshot (not a vanity-metric strip)
2. **People** — directory + drill-into-user side panel (real, scoped, multi-role aware)
3. **Roles & Capabilities** — read-only or editable role-cap matrix grounded in the registry
4. **Audit Trail** — RFC-style timeline with multi-axis filter + CSV/JSON export
5. **Sessions & Security Signals** — live sessions, anomalies, MFA coverage, lockouts
6. **Usage & License** — seat consumption, growth trend, invitation funnel

The current MetalDocs screen exposes only fragments of #1 and #2 and conflates online presence with activity. It ships dead `Ver todos` affordances, a single-role dropdown that silently destroys multi-role state, hardcoded area pickers, and no audit/session/role-matrix surfaces. Every leading SaaS admin console (Okta, Auth0, Datadog, Grafana, Vercel, Linear, Notion, GitHub Enterprise, Stripe) ships all six surfaces, and gates them with capability scopes much like MetalDocs's view-grade caps from ADR 0016.

The recommended IA at the end of this document maps each surface to the exact capability that gates it (`CapUserView`, `CapMembershipView`, `CapMetricsView`, `CapTaxonomyView`, `CapAuditRead`) and to RFC 9457 + paginated envelope contracts.

---

## 1. Industry Reference Analysis

For each reference, the relevant admin surface is dissected into: **anchor surfaces**, **drill-ins**, **gating signals**, and **takeaways for MetalDocs**.

### 1.1 Okta — Administrator Dashboard

Okta's admin console treats the dashboard as a **monitoring + decision surface**, not a CRUD page. The "View your org at a glance" widget surfaces overall usage; the **security monitoring widget** lets administrators review org-level metrics and user-reported suspicious activity at a glance; **HealthInsight** displays a graph of tasks completed for maintaining and increasing the org's security posture; agent status (LDAP, AD, on-prem) and Okta service status are first-class tiles. ([Okta admin dashboard docs](https://help.okta.com/oie/en-us/content/topics/dashboard/dashboard.htm), [Okta security monitoring](https://help.okta.com/en-us/content/topics/dashboard/monitor-org-security.htm), [Okta admin session protection](https://sec.okta.com/articles/protectingadminsessions/), [Okta recent activity / security events](https://help.okta.com/eu/en-us/content/topics/end-user/eu-recent-signin-activity.htm))

Anchor surfaces:
- Org-at-a-glance summary card (users / groups / apps / policies)
- HealthInsight security score with task completion graph
- Agent + service status tiles
- Security monitoring widget (suspicious activity, MFA failures)
- Reports center (sign-ins, MFA usage, lifecycle events)

Drill-ins: per-user "recent activity / security events" view exposing device, IP, location, and result for each sign-in. Admin sessions themselves are protected (re-authentication, step-up MFA) per the linked sec.okta.com guidance — admins cannot rely on a long-lived session to operate.

Takeaway for MetalDocs:
- Replace the "session" KPI cards with a **HealthInsight-style score** that aggregates: MFA coverage %, weak-password users, inactive >90d, admin overgrant. Each tile is a deep link into a remediation list.
- Surface a **device + IP + result** sub-table inside the user side panel — not just a flat "last activity" timestamp.

### 1.2 Auth0 — Security Center + Attack Protection

Auth0's Security Center is **real-time anomaly detection metrics + one-click mitigation**, co-located with the configuration screens. Each protection module ships a 7-day count: bot detections, suspicious IP throttles, brute-force blocks, breached-credential matches. The "Anomaly Detection" section was renamed "Attack Protection" and moved under a top-level **Security** node alongside MFA. ([Auth0 Security Center](https://auth0.com/docs/secure/security-center), [Auth0 attack protection](https://auth0.com/learn/anomaly-detection), [Auth0 anomaly detection docs](https://auth0.com/docs/anomaly-detection), [Auth0 community: improved security config UX](https://community.auth0.com/t/improved-experience-for-configuring-security-settings-in-our-dashboard/54385))

Anchor surfaces:
- "Tenant security pulse" — single-row strip with last-7d counts per detection module
- Per-module configure-and-observe panel (the metric and the mitigation toggle on the same screen)
- Breached-password detection card with provenance (HaveIBeenPwned-style)
- MFA enrollment funnel

Takeaway for MetalDocs:
- Adopt the **"metric next to mitigation"** pattern. Do not split "I see brute-force attempts" from "I lock the account" across two screens. On a QMS where document approval is signed, brute-force on signer accounts is a compliance event — surface it once.
- Pulse window of **7 days** is the right default rolling window for SaaS admin observability (Auth0, Vercel, Linear all use 7d or 90d).

### 1.3 Datadog — Org Settings, RBAC, Audit Trail, Usage

Datadog separates **Organization Settings** (admins manage users, groups, RBAC, keys, tokens, teams) from **Audit Trail** (read/write permission-gated, retention configurable, archivable to cloud storage). Audit Trail captures user logins, role changes, dashboard/monitor edits, and any privileged write. **Only users with Audit Trail Write permission can enable/disable Audit Trail; Audit Trail Read is required to view audit events.** ([Datadog Organization Settings](https://docs.datadoghq.com/account_management/org_settings/), [Datadog Audit Trail](https://docs.datadoghq.com/account_management/audit_trail/), [Datadog RBAC permissions](https://docs.datadoghq.com/account_management/rbac/permissions/), [Datadog blog: compliance & governance](https://www.datadoghq.com/blog/compliance-governance-transparency-with-datadog-audit-trail/))

Anchor surfaces:
- Organization Settings hub (Users / Groups / Roles / Service Accounts / API Keys / Audit Trail Settings)
- Audit Events Explorer (filterable timeline opening in a separate tab)
- RBAC permission matrix with three managed roles + custom roles
- Usage page with billing/usage break-down per product

Takeaway for MetalDocs:
- The right mental model is **"Audit Trail is itself a permission-gated capability"**. Map this to `CapAuditRead` (already in the registry per ADR 0016 §0.x context). Do not show audit even to `system_admin` without an explicit grant — it forces an explicit privilege footprint.
- Separate **Org Settings (configuration)** from **Observability (read-only signals)** in the IA. Datadog has a tab called "Audit Trail Settings" (configure retention/archive) *next to* the explorer; this is the right shape.

### 1.4 Grafana — Org Admin

Grafana scopes administration through three nested layers: **Server Admin** (across all orgs) → **Org Admin** (per organization) → **Folder/Dashboard Admin** (per resource). Service Accounts live under Administration → Users & Access → Service Accounts. Teams are the unit of permission grant. Audit logs are enterprise-only and require explicit enablement. ([Grafana Administration](https://grafana.com/docs/grafana/latest/administration/), [Grafana audit log](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/audit-grafana/), [Grafana service accounts](https://grafana.com/docs/grafana/latest/administration/service-accounts/), [Grafana roles and permissions](https://grafana.com/docs/grafana/latest/administration/roles-and-permissions/), [Grafana teams](https://grafana.com/docs/grafana/latest/administration/team-management/))

Anchor surfaces:
- Users & Access hub: Users / Teams / Service Accounts / Roles / Permissions
- Per-team membership editor with role-per-team rather than role-globally
- Audit logging (opt-in, log file or remote)
- Org preferences / branding

Takeaway for MetalDocs:
- Treat **Teams / Groups / Areas** as a first-class object. MetalDocs has Areas + Profiles as taxonomy entities — the admin panel should surface area membership the way Grafana surfaces team membership: a tab with each area, expanding to its approvers, signers, editors.
- Service accounts (long-lived machine identities) are a missing surface in MetalDocs today and should be on the roadmap for the Roles & Capabilities tab even if not in MVP.

### 1.5 Vercel — Team Admin, Audit Log, Usage

Vercel's team settings surface a **flat role set** (Owner / Member / Viewer / Billing, plus Enterprise-only Developer with project scope). Audit Logs are owner-only, 90-day window, **CSV export** is standard, and Enterprise customers can configure **real-time SIEM streaming** to Datadog/Splunk/S3/HTTP. ([Vercel audit logs](https://vercel.com/docs/audit-log), [Vercel access roles](https://vercel.com/docs/rbac/access-roles), [Vercel audit logs SIEM GA changelog](https://vercel.com/changelog/audit-logs-with-siem-integration-now-generally-available), [Vercel account management](https://vercel.com/docs/accounts))

Anchor surfaces:
- Team → Settings → Members (invitations, role assignment, removal)
- Team → Settings → Audit Log (filter + 90-day CSV export + SIEM stream)
- Usage page (per-resource consumption, soft limits, plan tier)
- Notifications routing

Takeaway for MetalDocs:
- Audit Log **CSV export is non-negotiable** — Brazilian-regulated QMS customers will need to hand evidence to auditors. JSON export second, SIEM stream third.
- "Invitation funnel" should be a visible state: sent → accepted → activated (first login).

### 1.6 Linear — Workspace Admin

Linear uses a clear two-level nav: **Workspace Settings** (visible to all members) and **Administration** (visible only to Admins/Owners). Administration contains **Members** (filter by role + status: Pending, Suspended, Left) and **Audit Log** (Owner-only, 90-day window, webhook stream for SIEM). ([Linear members and roles](https://linear.app/docs/members-roles), [Linear audit log](https://linear.app/docs/audit-log), [Linear audit log changelog](https://linear.app/changelog/2021-10-07-audit-log), [Linear workspaces](https://linear.app/docs/workspaces))

Anchor surfaces:
- Members table with multi-axis status filter (Active / Pending invite / Suspended / Left)
- Per-row role dropdown (single-role model — note: this is a *limitation* MetalDocs must NOT inherit, since MetalDocs is multi-role)
- Audit log with webhook stream toggle
- Workspace branding

Takeaway for MetalDocs:
- The **status filter is a missing dimension** in the current screen. Pending invitations, suspended users, and users who have left are not the same as active users and must not collapse into one list.
- Linear's single-role-dropdown anti-pattern is exactly what the current MetalDocs screen reproduces. The 8-role MetalDocs model (`system_admin, approver, author, editor, viewer, signer, area_admin, qms_admin`) is multi-assignable; the UI must be a checkbox group or multi-select, never a single dropdown.

### 1.7 Notion — Enterprise Admin Console

Notion's enterprise admin console organizes around **five categories**: General, People, Security, Data & Permissions, Analytics. The home screen is the bird's-eye view; audit logs live under Data & Compliance and can stream to Datadog/Panther/Splunk/Sumo. ([Notion audit log](https://www.notion.com/help/audit-log), [Notion enterprise admin](https://www.notion.com/help/category/enterprise-admin), [Notion org-level controls](https://www.notion.com/help/organization-level-controls), [Notion enterprise security provisions](https://www.notion.com/help/guides/notion-enterprise-security-provisions), [Notion audit log events (dev docs)](https://developers.notion.com/compliance/audit-log-events))

Anchor surfaces:
- Home dashboard (bird's-eye across all five categories)
- People (members, teamspaces, guest access)
- Security (SSO, SCIM, IP allowlist, session policy)
- Data & Permissions (audit log, exports, retention)
- Analytics (engagement, page metrics)

Takeaway for MetalDocs:
- The **five-category taxonomy is a strong IA pattern**: separate **People / Security / Data / Analytics / General**. Avoid putting "session policy" inside "Users" — these are different concerns.
- A QMS admin console should have a `Data & Compliance` section explicitly (export, retention, legal hold).

### 1.8 GitHub Enterprise — Site Admin Dashboard

The Site Admin Dashboard lets you manage users, organizations, and repositories. The audit log shows **180 days** of actions, filterable by qualifiers (action category, country code, organization) rather than free-text. The dashboard also surfaces inactive users and SSH key audits. ([GitHub site admin dashboard](https://docs.github.com/en/enterprise-server@3.5/admin/configuration/configuring-your-enterprise/site-admin-dashboard), [GitHub auditing users across enterprise](https://docs.github.com/en/enterprise-server@3.10/admin/managing-accounts-and-repositories/managing-users-in-your-enterprise/auditing-users-across-your-enterprise))

Anchor surfaces:
- Users panel (all users + admins + inactive users + SSH key audit)
- Org + repo panel
- Audit log query interface (qualifier-based, 180d retention)

Takeaway for MetalDocs:
- **Inactive-user surface as a separate, actionable list** (not a KPI) — "47 users inactive >90 days" is a remediation queue, not a stat.
- Qualifier-based filtering (`actor:`, `action:`, `resource:`) reads cleanly and is more accessible than free-text for non-power users.

### 1.9 Stripe — Team & Roles

Stripe ships **six fixed roles**: Owner, Administrator, Developer, Analyst, Support Specialist, View Only. No custom roles, no per-resource toggles. **Multiple roles can be assigned to one user and are additive**. Notable friction: no last-login in the team UI, no audit-log export from team management, API keys not auto-revoked on member removal. ([Stripe new roles & permissions blog](https://stripe.com/blog/new-roles-and-permissions-in-the-dashboard), [Stripe user roles docs](https://docs.stripe.com/get-started/account/teams/roles), [Stripe manage org access](https://docs.stripe.com/get-started/account/orgs/team?locale=en-GB))

Anchor surfaces:
- Team Settings → Members (invite, role assign, remove)
- Roles documentation page (read-only)
- Permission descriptions per role

Takeaway for MetalDocs:
- **Multiple additive roles is the right model and exactly what MetalDocs already does**. The UI must not silently destroy multi-role state. Stripe's documentation explicitly says "if you assign a user multiple roles, they're assigned all the permissions of each individual role."
- **Avoid Stripe's gaps**: surface last-login per user, surface session/key revocation on member removal, surface audit export.

### 1.10 Cross-reference: SIEM / Audit Streaming Patterns

Patterns harvested from the audit-log-design guides cited below: enterprise filters that match how humans investigate are **actor / action / resource / date range / result / IP**. Exports support **CSV + JSON**. SIEM forwarding is per-tenant, per-destination, with independent filters and formatting (Splunk JSON, Datadog JSON, ECS, OCSF). Delivery is **at-least-once with exponential backoff and a persistent queue**. A **test event endpoint** lets customers validate integration before enabling production. ([AverageDevs — designing audit logs for SaaS](https://www.averagedevs.com/blog/audit-logs-saas-compliance-trust), [AuditKit SIEM integration guide](https://auditkit.dev/blog/siem-integration-audit-logs), [WorkOS developer's guide to audit logs / SIEM](https://workos.com/blog/the-developers-guide-to-audit-logs-siem))

---

## 2. Recurring Building Blocks (Pattern Catalog)

Synthesized across all references. Each block is a reusable UI primitive the MetalDocs admin panel should ship.

| # | Pattern | Description | Where seen |
|---|---|---|---|
| P1 | **KPI strip (≤ 7 tiles, ≤ 1 row)** | Top-of-page health tiles. Each tile is a deep-link, not a vanity number. | Okta, Datadog, Notion, Vercel |
| P2 | **HealthInsight score** | Composite security/posture score with task-completion graph and remediation deep-links. | Okta, Auth0 |
| P3 | **Presence stream (real)** | WebSocket-backed online status with TTL heartbeat (not polling, not faked). Multi-tab safe. | Figma, Slack, Linear, Google Docs |
| P4 | **Directory table + side panel** | Sortable, filterable, paginated user table. Row click opens a right-side drawer with full detail. | Okta, Auth0, Linear, Vercel, Notion |
| P5 | **Audit timeline with multi-axis filter** | Reverse-chronological, filterable by actor/action/resource/date/result/IP, exportable to CSV+JSON, optionally SIEM-streamed. | Datadog, Vercel, Linear, Notion, GitHub |
| P6 | **Session manager** | Per-user active sessions with device/IP/location/last-active, revocable individually or in bulk. | Okta, Auth0, Google Workspace |
| P7 | **Role/Capability matrix** | Subjects × Resources × Actions table, read-only or editable. | Datadog, AWS IAM, Grafana, Stripe |
| P8 | **Security signals panel** | Last-7-day counts: failed logins, brute-force, MFA failures, breached creds, suspicious IPs. Each metric next to its mitigation toggle. | Auth0, Okta |
| P9 | **Usage / license panel** | Seats used vs limit, growth trend, soft/hard limits, plan tier. | Vercel, Datadog, Linear |
| P10 | **Invitation funnel** | Sent → Accepted → Activated states with per-row resend / revoke. | Linear, Notion, Vercel |
| P11 | **Anomaly cards** | Per-event mini-cards on the overview ("3 brute-force attempts blocked from IP X in the last hour — investigate"). | Auth0 Security Center |
| P12 | **Drill-into-user side panel** | Right-side drawer with tabs: Profile / Roles / Sessions / Recent Activity / Security / Memberships. | Okta, Auth0, Linear |
| P13 | **Bulk actions** | Multi-select checkbox column with bulk role-grant, deactivate, force-logout, password-reset. | Okta, Auth0, Datadog, Notion |
| P14 | **Scope-bounded RBAC view** | Admin only sees the slice of users/areas they govern (e.g. `area_admin` sees only their area's members). | Grafana (org admin), AWS IAM (scoped) |
| P15 | **Status filters (not status badges only)** | First-class filter facet: Active / Pending / Suspended / Left / Locked. | Linear, Vercel |
| P16 | **Last-login + IP + device per user** | Always-on column in the directory and always-on tab in the side panel. | Okta, Auth0, Vercel |
| P17 | **"Metric next to mitigation" pattern** | Don't split observation from action. The brute-force count and the lockout toggle live on one screen. | Auth0 |
| P18 | **Permission-gated tab visibility** | Whole tabs disappear (not just rows) when the actor lacks the cap. The route renders an explicit 403 boundary, not an empty page. | Datadog, Grafana, Notion |

---

## 3. Information Architecture for MetalDocs `/admin`

The recommended IA is **six top-level tabs**, each owned by a distinct view-grade capability from ADR 0016. Tabs are hidden — not just disabled — when the actor lacks the gating cap. The Overview tab is the only one visible to all roles holding `CapUserView` OR `CapMembershipView`.

### 3.1 Overview (default landing)

**Gating cap:** any of `CapUserView`, `CapMembershipView`, `CapMetricsView`.
**Purpose:** decision-oriented snapshot. Not a vanity dashboard.

Primary surfaces (above the fold):
- HealthInsight score (P2) — composite of MFA coverage %, weak-password users, inactive users, admin overgrant, MFA-not-enrolled-on-signer-role.
- Security signals strip (P8) — last-7d counts: failed logins, brute-force blocks, suspicious IPs, breached-credential matches, lockouts.
- Anomaly cards (P11) — up to 3 actionable cards. "X failed signer logins on doc Y" → deep link.
- Presence card (P3) — count + live mini-list of online users (real WebSocket presence, not poll).

Secondary (below the fold):
- Recent privileged actions (last 20 from audit) — deep link to Audit tab.
- Pending invitations widget (P10).

What does NOT belong here: full user CRUD, role matrix, raw activity dump.

### 3.2 People

**Gating cap:** `CapUserView`.
**Purpose:** directory + drill-in. Replaces the current `ManagedUsersSection`.

Primary surface: a paginated, sortable, filterable directory (P4):
- Columns: avatar, displayName, email, roles (chips), area memberships (chips), status (Active/Pending/Suspended/Locked/Left — P15), lastLogin, lastIP, MFA enrolled (Y/N), actions.
- Multi-axis filter facets: status, role, area, MFA status, last-login window.
- Bulk actions (P13): grant role, revoke role, deactivate, force-logout, send password reset.
- Search by name / email.

Drill-in (side panel — P12, P4) opens on row click with six tabs:
1. **Profile** — display name, email, language, timezone, created-at, last-modified-by.
2. **Roles** — multi-select **checkbox group** showing all 8 canonical roles (`system_admin, approver, author, editor, viewer, signer, area_admin, qms_admin`). Never a single dropdown. Shows the role-cap implication ("granting `approver` adds: cap.approval.manage, cap.document.view, …").
3. **Memberships** — table of area memberships with role-per-area (Grafana-style).
4. **Sessions** — active sessions list (P6): device, IP, location, last-active, revoke button. Bulk revoke all.
5. **Recent Activity** — last 50 audit entries scoped to this user (deep link to full Audit tab with `actor:userId` filter pre-applied).
6. **Security** — MFA enrolment, last password change, failed login count last 7d, lockout status, "force password reset on next login" toggle.

What does NOT belong here: hardcoded area dropdowns (areas come from `/api/v1/taxonomy/areas` gated by `CapTaxonomyView`); single-role dropdowns; the legacy "managed users" naming.

### 3.3 Roles & Capabilities

**Gating cap:** `CapMembershipView` (read) + `CapMembershipManage` (edit). Edit may be `system_admin`-only in MVP.
**Purpose:** make the cap registry **legible**. Today it lives only in code and SQL — admins cannot see what `approver` actually grants.

Primary surface: a **subjects × resources × actions** matrix (P7) where:
- Rows: the 8 canonical roles.
- Columns: capability domains (Document, Approval, Taxonomy, User, Membership, Audit, Metrics).
- Cells: dots for granted view-grade / manage-grade / submit-grade caps. Hover/click expands to the exact cap constant (`cap.document.view`, `cap.approval.manage`, …).
- A read-only mode in MVP. Editing flows ship in a later milestone after the cap-registry mutation API exists.

Secondary surface: per-role detail panel listing the full cap list, the routes those caps gate, and the user count holding that role.

What does NOT belong here: invent new roles in the UI (the 8 canonical roles are seeded; new roles require a migration). No free-text "create custom role" affordance until the backend supports it.

### 3.4 Audit Trail

**Gating cap:** `CapAuditRead`.
**Purpose:** compliance + investigation. RFC-shaped timeline.

Primary surface: a reverse-chronological timeline (P5) with the following filter facets (URL-state-persisted per `web/patterns.md` URL-as-state rule):
- Actor (user picker; default empty)
- Action category (login, logout, role-grant, role-revoke, doc-create, doc-publish, approval-submit, approval-decide, signature, settings-change, session-revoke)
- Resource type + ID (document, approval-route, area, user, role, membership)
- Date range (preset chips: 24h / 7d / 30d / 90d; custom range)
- Result (success / failure)
- IP / country (when telemetry attaches it)

Per-row: timestamp (BRT, `pt-BR` locale), actor avatar + name, action chip, resource link, result indicator, expand for full diff/payload (RFC 9457 problem details for failures).

Actions: CSV export, JSON export, copy permalink to filtered view. SIEM streaming config lives in a separate "Audit Settings" sub-tab gated by `CapAuditWrite` (future).

What does NOT belong here: free-text-only search. Use qualifier filters (P5) — GitHub Enterprise pattern — because they are auditable and bookmarkable.

### 3.5 Sessions & Security

**Gating cap:** `CapUserView` (read sessions) + `CapMembershipManage` for revoke. Org-wide security policy edits are `system_admin` only.
**Purpose:** the actionable half of the security signals strip on Overview.

Primary surfaces:
- **Active sessions table** (P6) across all users — columns: user, device, IP, country, started-at, last-active, MFA at login (Y/N). Bulk revoke. Filter by user, by IP, by country, by MFA-Y/N.
- **Security signals detail** (P8 expanded) — last-7d and last-30d series for: failed logins, brute-force, suspicious IPs, breached-credential matches, lockouts. Each metric next to its mitigation (P17): toggle account-lockout policy, toggle suspicious-IP throttling, configure MFA requirement scope.
- **MFA coverage panel** — % enrolled by role. Highlight roles where MFA is critical (`signer`, `approver`, `system_admin`, `qms_admin`).
- **Lockouts list** — currently-locked accounts with unlock action.

What does NOT belong here: an "online users" widget (that lives on Overview). Sessions ≠ presence.

### 3.6 Usage & License

**Gating cap:** `CapMetricsView`.
**Purpose:** seat consumption, system load, plan tier.

Primary surfaces:
- Seats: used / total, growth trend (last 12 weeks), projected exhaustion date.
- Active vs licensed users (active = logged in in last 30 days).
- Storage / document volume snapshot (read from existing metrics where exposed).
- API request volume (if MetalDocs exposes a public API surface in scope).
- Plan tier + entitlement matrix (which features are on/off).

What does NOT belong here: per-user billing detail (Brazilian QMS is org-licensed, not per-seat-billed at MVP). Skip until product confirms billing model.

### 3.7 General / Org Settings (optional, deferred)

Notion-style fifth bucket: branding, language defaults, retention policy, SSO/SCIM config (future), legal hold (future). Out of MVP scope but reserve the IA slot.

---

## 4. Capability Model Alignment (ADR 0016)

Mapping each proposed surface to the exact cap that gates it. Anchored in `wiki/decisions/0016-view-grade-capabilities.md` and the 8 canonical IAM roles.

| Tab | Surface | Gating cap (read) | Gating cap (write) | Roles that hold the read cap |
|---|---|---|---|---|
| Overview | HealthInsight score | `CapMembershipView` ∪ `CapUserView` | n/a | all roles for membership view; `system_admin` + `area_admin` for user view |
| Overview | Security signals strip | `CapMetricsView` | `CapMetricsManage` (future) | `system_admin` only |
| Overview | Anomaly cards | `CapMetricsView` | n/a | `system_admin` only |
| Overview | Presence card | `CapUserView` | n/a | `system_admin`, `area_admin` |
| People | Directory table | `CapUserView` | `CapUserManage` | `system_admin`, `area_admin` |
| People | Drill-in → Roles tab | `CapUserView` | `CapUserManage` + `CapMembershipManage` | `system_admin`, `area_admin` |
| People | Drill-in → Memberships tab | `CapMembershipView` | `CapMembershipManage` | all roles read; `system_admin` write |
| People | Drill-in → Sessions tab | `CapUserView` | `CapUserManage` (revoke) | `system_admin`, `area_admin` |
| People | Drill-in → Recent Activity | `CapAuditRead` | n/a | `system_admin` (today) |
| People | Drill-in → Security | `CapUserView` | `CapUserManage` | `system_admin`, `area_admin` |
| Roles & Capabilities | Cap matrix (read) | `CapMembershipView` | n/a | all roles |
| Roles & Capabilities | Cap matrix (edit) | n/a | `CapMembershipManage` | `system_admin` (future) |
| Audit Trail | Timeline | `CapAuditRead` | n/a | `system_admin` |
| Audit Trail | Export CSV/JSON | `CapAuditRead` | n/a | `system_admin` |
| Audit Trail | SIEM stream config | n/a | `CapAuditWrite` (future) | `system_admin` |
| Sessions & Security | Active sessions table | `CapUserView` | `CapUserManage` | `system_admin`, `area_admin` |
| Sessions & Security | Security signals detail | `CapMetricsView` | `CapMetricsManage` (future) | `system_admin` |
| Sessions & Security | MFA coverage panel | `CapUserView` | `CapUserManage` | `system_admin`, `area_admin` |
| Sessions & Security | Lockouts list | `CapUserView` | `CapUserManage` | `system_admin`, `area_admin` |
| Usage & License | Seats panel | `CapMetricsView` | n/a | `system_admin` |
| Usage & License | Storage / API volume | `CapMetricsView` | n/a | `system_admin` |
| Usage & License | Plan tier | `CapMetricsView` | n/a | `system_admin` |

**`area_admin` scope-bounded behavior** (P14): when an `area_admin` opens People, the directory is filtered server-side to users with membership in any of the admin's areas. The same filter applies to Sessions & Security (only sessions of users they govern). The Roles & Capabilities tab is read-only for `area_admin`. The Audit Trail tab is hidden until/unless `CapAuditRead` is granted to `area_admin` (today: not granted; product decision pending).

**Tab visibility rule** (P18): tabs are removed from the tab strip when the actor lacks the gating cap, not greyed-out. The route renders an explicit 403 boundary via the existing `WorkspaceDataState` error pattern, not an empty/dead page. This matches Datadog/Notion/Grafana behavior and avoids the current screen's "I see a tab I can't use" footgun.

---

## 5. Data Shape per Surface (API Contract Hints)

All responses follow MetalDocs's existing contract-first OpenAPI pattern (`wiki/architecture/api-contract.md`, `wiki/architecture/api-design-system.md`). Errors use RFC 9457 problem-details envelopes. Lists are paginated with `{ items: T[], pagination: { page, pageSize, total } }`.

### 5.1 Overview surfaces

**`GET /api/v1/admin/overview`** — cap `CapMembershipView` ∪ `CapUserView`.

```
{
  healthInsight: { score: 0..100, breakdown: { mfaCoverage, weakPasswords, inactiveOver90d, adminOvergrant, signerWithoutMfa } },
  signals7d:     { failedLogins, bruteForceBlocks, suspiciousIps, breachedCreds, lockouts },
  anomalies:     [{ id, severity, summary_ptBR, deepLink, observedAt }],
  presence:      { onlineCount, sample: [{ userId, displayName, since }] }
}
```

**`GET /api/v1/admin/presence/stream`** — WebSocket, cap `CapUserView`. Heartbeat 30s, TTL 60s in Redis (P3). Emits `presence.join`, `presence.leave`, `presence.heartbeat`.

### 5.2 People

**`GET /api/v1/admin/users`** — cap `CapUserView`. Paginated.

Query: `?status=active|pending|suspended|locked|left&role=…&area=…&mfa=yes|no&lastLoginWindow=24h|7d|30d|90d&q=…&page=&pageSize=`.

```
{
  items: [{
    userId, displayName, email,
    roles: ["approver","editor"],
    areaMemberships: [{ areaId, areaCode, roleInArea }],
    status: "active|pending|suspended|locked|left",
    mfaEnrolled: bool,
    lastLoginAt, lastLoginIp, lastLoginCountry,
    createdAt, updatedAt
  }],
  pagination: { page, pageSize, total }
}
```

**`GET /api/v1/admin/users/{userId}`** — cap `CapUserView`. Same shape, single item, plus `recentActivityCount7d`.

**`PUT /api/v1/admin/users/{userId}/roles`** — cap `CapUserManage`. Body `{ roles: string[] }`. **Atomic multi-role replace** — never derive from a single-value dropdown.

**`POST /api/v1/admin/users/{userId}/sessions/revoke`** — cap `CapUserManage`. Body `{ sessionIds: string[] | "all" }`.

**`GET /api/v1/admin/users/{userId}/sessions`** — cap `CapUserView`.

```
{
  items: [{ sessionId, deviceLabel, userAgent, ip, country, startedAt, lastActiveAt, mfaAtLogin, current: bool }],
  pagination: { … }
}
```

### 5.3 Roles & Capabilities

**`GET /api/v1/admin/role-capabilities`** — cap `CapMembershipView`.

```
{
  roles: [{
    role: "approver",
    capabilities: ["cap.document.view","cap.approval.manage", … ],
    userCount: 47
  }],
  capabilities: [{
    constant: "cap.document.view",
    description_ptBR: "Visualizar documentos",
    grantedToRoles: ["approver","editor","viewer", … ],
    gatesRoutes: ["GET /api/v1/documents", … ]
  }]
}
```

### 5.4 Audit Trail

**`GET /api/v1/admin/audit-events`** — cap `CapAuditRead`. Paginated, sorted desc by `occurredAt`.

Query: `?actor=&action=&resourceType=&resourceId=&from=&to=&result=success|failure&ip=&country=&page=&pageSize=`.

```
{
  items: [{
    eventId,
    occurredAt,           // ISO 8601 with TZ; FE renders in pt-BR / America/Sao_Paulo
    actor: { userId, displayName },
    action,               // canonical verb, e.g. "approval.decide"
    resource: { type, id, label },
    result: "success|failure",
    ip, country, userAgent,
    diff?: { before, after },
    problem?: RFC9457    // when result=failure
  }],
  pagination: { page, pageSize, total }
}
```

**`GET /api/v1/admin/audit-events/export`** — cap `CapAuditRead`. Query params identical. `Accept: text/csv` or `application/json`. Server streams the response.

### 5.5 Sessions & Security

**`GET /api/v1/admin/sessions`** — cap `CapUserView`. Org-wide active session list. Paginated, filterable by user, IP, country, MFA.

**`GET /api/v1/admin/security/signals`** — cap `CapMetricsView`. Time-series for the last-7d and last-30d windows.

```
{
  windows: ["7d","30d"],
  series: {
    failedLogins:      [{ bucketStart, count }],
    bruteForceBlocks:  [...],
    suspiciousIps:     [...],
    breachedCreds:     [...],
    lockouts:          [...]
  }
}
```

**`GET /api/v1/admin/security/mfa-coverage`** — cap `CapUserView`. `{ byRole: { roleName: { enrolledCount, totalCount, percent } } }`.

**`GET /api/v1/admin/security/lockouts`** — cap `CapUserView`. List of currently locked users.

### 5.6 Usage & License

**`GET /api/v1/admin/usage`** — cap `CapMetricsView`. Seats, growth, storage, API volume.

### 5.7 Error envelope

Every failure returns RFC 9457 problem details:

```
{
  type:    "https://metaldocs.example/problems/insufficient-capability",
  title:   "Insufficient capability",
  status:  403,
  detail:  "Actor lacks cap.audit.read",
  instance:"/api/v1/admin/audit-events",
  capRequired: "cap.audit.read",
  actorRoles:  ["editor"]
}
```

Frontend renders `detail` in `pt-BR` via the i18n layer (server returns the canonical English string + a code; FE maps the code to a Portuguese message).

---

## 6. Non-Negotiables

1. **Accessibility — WCAG 2.2 AA** ([WebAIM checklist](https://webaim.org/standards/wcag/checklist), [Deque WCAG 2.2 AA checklist](https://media.dequeuniversity.com/en/docs/web-accessibility-checklist-wcag-2.2.pdf)):
   - All controls keyboard-reachable, with visible focus rings (criterion 2.4.11 Focus Not Obscured, 2.4.13 Focus Appearance — new in 2.2).
   - Data tables: `<th scope>` on header cells, `caption` describing the table, row/column header association. Pagination + sort controls must be announced via `aria-live` polite region when state changes.
   - Side panel: traps focus when open, restores focus to the row trigger on close, `Esc` closes.
   - Tabs: roving tabindex, `aria-selected`, `aria-controls`, arrow-key navigation. No reliance on color alone (WCAG 1.4.1 Use of Color) for status — pair every status color with an icon and a label.
   - Color contrast ≥ 4.5:1 for body text, ≥ 3:1 for large text and UI components (1.4.11).
   - Timing-adjustable: session timeout must offer an extend prompt before logout (2.2.1).
   - Drag-and-drop affordances (if any in bulk-select) must have a pointer-based alternative (2.5.7 — new in 2.2).
2. **Brazilian Portuguese** — all visible copy in `pt-BR`. Date formatting via `Intl.DateTimeFormat("pt-BR", { dateStyle, timeStyle })`, timezone `America/Sao_Paulo`. Number formatting via `Intl.NumberFormat("pt-BR")`. Never embed English fallback in user-visible strings — use the i18n layer (errors come from server as codes + English `title`, translated FE-side).
3. **Dark + Light readiness** — design tokens drive both palettes. Status colors (success, warning, danger, info) must hold the 4.5:1 contrast threshold in both themes. No `oklch(98% 0 0)` hard-coded white backgrounds in components; use `var(--color-surface)`.
4. **No decorative / dead UI** — every link, button, and badge must be hooked to a real action. No `Ver todos` that goes nowhere. No fake KPIs. No mocked counts.
5. **No hardcoded dropdowns** — areas, profiles, roles come from real APIs (`/api/v1/taxonomy/areas`, `/api/v1/taxonomy/profiles`, role registry endpoint).
6. **Audit trail filter + export** — CSV export is MVP, JSON is fast follow, SIEM streaming is later. Server streams the export; FE never accumulates the full payload in memory.
7. **Real presence** — WebSocket-backed online status with TTL heartbeat (Redis-side). Polling and faking are forbidden.
8. **Permission-gated tab visibility** — tabs are removed from the tab strip when the cap is missing. Routes render a 403 boundary, never an empty container.
9. **Multi-role aware UI** — checkbox group or multi-select for roles. Never a single-value dropdown. The server route for role assignment is a **replace** call with the full role set, not an add/remove that races.
10. **Optimistic mutations + rollback** — per the repo's existing mutation factory pattern (recent commit `6fce76d0e`). Show toast on rollback. Never silently swallow.
11. **URL-as-state** — all filter facets in Audit and People persist via search params (per `web/patterns.md`).

---

## 7. Anti-Patterns (Explicit Don't-Ship List)

Distilled from the current `AdminCenterView.tsx`, the references, and the prompt's hint list.

| # | Anti-pattern | Why it's wrong | What to do instead |
|---|---|---|---|
| A1 | Dead `Ver todos` link | No destination. Lies to the user about navigation. | Link to a real filtered view, or remove the link. |
| A2 | Hardcoded `area` / `profile` dropdowns | Bypasses the taxonomy authority; goes stale; breaks `CapTaxonomyView` audit. | Fetch from `/api/v1/taxonomy/{areas,profiles}` with a typed hook. |
| A3 | Single-role dropdown on a multi-role user | Silently destroys 7/8 role assignments when you click "save". | Checkbox group or multi-select; server PUT replaces the full set atomically. |
| A4 | "Session" KPI cards that aren't actionable | Vanity metric. A number with no deep-link is a decoration. | Every tile deep-links to a filtered list. If it can't, kill the tile. |
| A5 | Conflating presence with activity | Online = WebSocket session alive. Activity = audit event. They are different concepts and lying about either hurts compliance. | Two separate surfaces. Presence card (real WS) + Audit timeline (truth log). |
| A6 | "Recent activity" as a free-floating list | No filters, no pagination, no export. Useless for compliance. | Move into the Audit Trail tab with full filter facets and CSV export. |
| A7 | Three KPI tiles labeled "online / last activity / total users" | Tells the admin nothing about the system's posture. None of these drive a decision. | Replace with HealthInsight score + security signals strip. |
| A8 | Inline string-based activity classification (`if action.includes('login')`) | Fragile, locale-specific, silently mis-labels. | Server emits canonical action constants; FE maps via a typed enum. |
| A9 | One-tab page with everything dumped together | Cognitive overload, no scope boundary, breaks cap-gated tab visibility. | Six tabs, each cap-gated. |
| A10 | "Managed users" terminology | Implies a second-class entity. All users are users. | Rename to "People" / "Usuários" (`pt-BR`). |
| A11 | Force-mounting CRUD form to the page bottom | Mobile-hostile, keyboard-trap-prone, no focus management. | CRUD lives inside the drill-in side panel. |
| A12 | Polling for presence | Wastes bandwidth, has 30-60s lag, fights with browser sleep. | WebSocket with heartbeat. |
| A13 | No CSV export on audit | Brazilian QMS auditors need evidence in CSV. | Ship CSV in MVP, JSON in v2. |
| A14 | Showing `system_admin` users + protected roles to all admins | Privacy leak; `area_admin` shouldn't see the system_admin list. | Server-side scope filter (P14). |
| A15 | Free-text search as the only audit filter | Not auditable, not bookmarkable, not accessible. | Qualifier-based facet filter (GitHub pattern). |
| A16 | "Always-on" `system_admin` impersonation | Long-lived elevated session. | Step-up auth + time-boxed impersonation session with explicit reason field, audited (per [Yaro Labs impersonation guide](https://yaro-labs.com/blog/user-impersonation-tool-saas)). |
| A17 | Showing raw cap constants without translation | `cap.approval.manage` is meaningless to a non-engineer. | Every cap has a `description_ptBR`. |
| A18 | Letting the cap matrix be editable without a migration story | The 8 canonical roles are seeded in SQL; UI-mutated grants will drift from `db/reference-data/`. | Read-only in MVP. Edit lands when the curated baseline supports round-trip from UI to migration. |

---

## 8. Reference Designs (Visual Inspiration)

Each link with a one-line rationale. These are *visual references* — the IA above is the source of truth.

1. **[Dribbble — Audit Logs UI by Divyansh Pandey](https://dribbble.com/tags/audit-logs)** — clean side-panel pattern for audit-event expansion; useful as a layout reference for the Audit Trail row-expand state.
2. **[Dribbble — Timeline filtering UI by Mike Hince](https://dribbble.com/shots/15449152-Timeline-filtering-UI-Design)** — qualifier-style filter chips above a timeline; matches the P5 + GitHub Enterprise pattern.
3. **[Dribbble — Dark Mode Panel, Diagram Management Dashboard by Diana Larussa](https://dribbble.com/shots/25972017-Dark-Mode-Panel-Diagram-Management-Dashboard-User-Flows-SaaS)** — disciplined dark-mode SaaS layout with intentional hierarchy; useful as a counterexample to default-template look.
4. **[Dribbble — Dark Mode Admin Dashboard by Dmitry Sergushkin](https://dribbble.com/shots/21881540-Dark-Mode-Admin-Dashboard)** — bento-style admin overview; informs the Overview tab composition (HealthInsight + signals strip + anomaly cards).
5. **[Dribbble — Audit Dashboard collection](https://dribbble.com/search/audit-dashboard)** — broader survey; useful for status-chip and filter-pill treatments.
6. **[Auth0 Security Center (live product)](https://auth0.com/docs/secure/security-center)** — gold standard for "metric next to mitigation" composition (P17).
7. **[Datadog Audit Trail (live docs)](https://docs.datadoghq.com/account_management/audit_trail/)** — gold standard for cap-gated audit UI and SIEM integration affordances.
8. **[Linear Audit Log (live docs)](https://linear.app/docs/audit-log)** — minimal, focused, owner-only — informs the visual restraint we want.
9. **[Notion Enterprise Admin Console](https://www.notion.com/help/category/enterprise-admin)** — the five-category nav model the IA borrows.
10. **[Setproduct — SaaS Dashboard examples](https://www.setproduct.com/blog/saas-dashboard-examples)** — survey of professional SaaS dashboard compositions; counter-reference for the "no template look" rule.

---

## 9. Implementation Notes

These are scoping hints for the engineer who picks this up, not prescription.

- **Tab routing** uses the existing `createBrowserRouter` + per-feature `routes.tsx` pattern. The `/admin` parent route nests six child routes. Per `wiki/architecture/frontend-structure.md`, the feature lives under `frontend/apps/web/src/features/admin/` (not `iam/` — the new feature is broader than IAM).
- **Server state** via TanStack Query with the existing optimistic-mutation factory (commit `6fce76d0e`). Query keys: `["admin","overview"]`, `["admin","users",filterParams]`, `["admin","audit",filterParams]`, etc. Cache invalidation on mutate scoped to the touched key prefix.
- **Generated types** from `lib/api-types/` after the OpenAPI spec adds the `/admin/*` paths. No hand-rolled types.
- **CSS Modules + design tokens** — every status color resolves to a token (`--color-status-success`, `--color-status-danger`, `--color-status-warning`, `--color-status-info`) defined in both light and dark themes.
- **WebSocket presence** — the backend ships a new presence server (Redis TTL pattern from [OneUptime presence guide](https://oneuptime.com/blog/post/2026-02-02-websocket-presence-detection/view)). Cap-gated; only `CapUserView` holders can subscribe. Heartbeat 30s, TTL 60s.
- **CSV export** is a server-side stream (Content-Disposition attachment). FE issues a regular GET with `Accept: text/csv` — no fetch-and-blob accumulation.
- **Step-up auth** on dangerous actions (impersonation, mass force-logout, role revoke on `system_admin`) — re-prompt for password or MFA. Pattern from [Okta admin session protection](https://sec.okta.com/articles/protectingadminsessions/).
- **Migration path off the current screen**: keep the old `AdminCenterView.tsx` mounted at a legacy route only for the duration of the rollout; new route is `/admin` with the six-tab IA. Delete the old route + module file once the new screen ships green QA.

---

## 10. Recommended IA for MetalDocs `/admin` (Concrete Tab Tree)

```
/admin                                      ← parent layout; redirects to /admin/overview
│
├── /admin/overview                         [gate: CapMembershipView ∪ CapUserView]
│     ├─ HealthInsight score card           (composite posture metric)
│     ├─ Security signals strip — 7d        (failed logins, brute-force, susp. IPs, breached, lockouts)
│     ├─ Anomaly cards                      (≤3 actionable cards, deep-link)
│     ├─ Presence card                      (real WebSocket presence, count + sample)
│     ├─ Pending invitations widget         (Sent → Accepted → Activated funnel)
│     └─ Recent privileged actions (last 20) → deep-link to /admin/audit
│
├── /admin/people                           [gate: CapUserView]
│     ├─ Toolbar: search + filter facets (status, role, area, MFA, last-login window)
│     ├─ Bulk actions bar (grant/revoke role, deactivate, force-logout, reset password)
│     ├─ Directory table (paginated, sortable, status-aware)
│     ├─ /admin/people/:userId              (row click → right-side drawer)
│     │     ├─ Tab: Perfil
│     │     ├─ Tab: Funções        (checkbox group, multi-role aware)
│     │     ├─ Tab: Áreas          (memberships, role-per-area)
│     │     ├─ Tab: Sessões        (active sessions, revoke individually or all)
│     │     ├─ Tab: Atividade      (last 50; deep-link to Audit pre-filtered)
│     │     └─ Tab: Segurança      (MFA, last pwd change, failed logins, lockout, force reset)
│     └─ /admin/people/invite              (modal: invite by email; choose roles + areas)
│
├── /admin/roles                            [gate: CapMembershipView; edit: CapMembershipManage]
│     ├─ Matrix view                        (8 roles × cap-domain columns; dot per granted cap)
│     ├─ Per-role detail panel              (full cap list + gated routes + user count)
│     └─ Per-capability detail panel        (description_ptBR + roles holding it + routes gated)
│
├── /admin/audit                            [gate: CapAuditRead]
│     ├─ Filter facets (actor, action, resource, date, result, IP, country)
│     ├─ Timeline (reverse-chronological)
│     ├─ Row expand → full payload (RFC 9457 details on failure)
│     ├─ Export CSV / JSON
│     └─ /admin/audit/settings              [gate: CapAuditWrite — future]
│           ├─ Retention config
│           └─ SIEM stream destinations (Datadog/Splunk/S3/HTTP)
│
├── /admin/sessions                         [gate: CapUserView; revoke: CapUserManage]
│     ├─ Active sessions table (org-wide, filterable)
│     ├─ Security signals detail (7d + 30d series, metric-next-to-mitigation)
│     ├─ MFA coverage panel (% by role; signer/approver/system_admin/qms_admin highlighted)
│     └─ Lockouts list (current locked accounts, unlock action)
│
└── /admin/usage                            [gate: CapMetricsView]
      ├─ Seats panel (used/total, growth, projected exhaustion)
      ├─ Active vs licensed (30d active)
      ├─ Storage / document volume snapshot
      ├─ API request volume (when applicable)
      └─ Plan tier + entitlement matrix
```

**Visibility rules:**

- An actor with no `Cap*View` cap gets a 403 boundary on `/admin` (no empty page).
- An `area_admin` sees Overview, People (scoped to their areas), Sessions (scoped), Usage (scoped if product confirms area-scoped metrics). Roles, Audit hidden unless explicitly granted.
- A `viewer` / `author` / `editor` / `approver` / `signer` has no `CapUserView` and no `CapAuditRead`; `/admin` is hidden from the workspace nav.
- A `system_admin` sees all six tabs.
- A `qms_admin` sees Overview, People, Audit (gated by `CapAuditRead` — product to confirm grant), Roles (read-only).

**Tab strip order** (left → right): Overview → People → Roles → Audit → Sessions → Usage. Order optimizes for: most-frequent decision surface first, most-administrative surface last.

**Mobile / narrow viewport:** tabs collapse into a horizontal scroller above the content; side drawer becomes a full-screen sheet.

**Empty states:** every list has a real empty state ("Nenhum evento no período selecionado. Ajuste o filtro ou amplie o intervalo."). Never a blank panel.

**Loading states:** skeleton rows matching the final table shape — never a centered spinner over an empty page (per `web/patterns.md` stale-while-revalidate guidance).

**Error states:** RFC 9457 problem-detail rendered with a `pt-BR` translation, a "Tentar novamente" button, and (when applicable) a deep-link to the cap docs that would unblock the actor.

---

## Sources

### Industry references (product docs)

- [Okta — Administrator Dashboard](https://help.okta.com/oie/en-us/content/topics/dashboard/dashboard.htm)
- [Okta — Monitor your org's security](https://help.okta.com/en-us/content/topics/dashboard/monitor-org-security.htm)
- [Okta — Protecting Administrative Sessions](https://sec.okta.com/articles/protectingadminsessions/)
- [Okta — Recent Activity & Security Events](https://help.okta.com/eu/en-us/content/topics/end-user/eu-recent-signin-activity.htm)
- [Okta — Session management (developer)](https://developer.okta.com/docs/concepts/session/)
- [Auth0 — Security Center](https://auth0.com/docs/secure/security-center)
- [Auth0 — Attack Protection (anomaly detection)](https://auth0.com/learn/anomaly-detection)
- [Auth0 — Attack Protection docs](https://auth0.com/docs/anomaly-detection)
- [Auth0 community — improved security configuration UX](https://community.auth0.com/t/improved-experience-for-configuring-security-settings-in-our-dashboard/54385)
- [Datadog — Organization Settings](https://docs.datadoghq.com/account_management/org_settings/)
- [Datadog — Audit Trail](https://docs.datadoghq.com/account_management/audit_trail/)
- [Datadog — RBAC permissions](https://docs.datadoghq.com/account_management/rbac/permissions/)
- [Datadog blog — compliance & governance via Audit Trail](https://www.datadoghq.com/blog/compliance-governance-transparency-with-datadog-audit-trail/)
- [Grafana — Administration overview](https://grafana.com/docs/grafana/latest/administration/)
- [Grafana — Audit a Grafana instance](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/audit-grafana/)
- [Grafana — Service accounts](https://grafana.com/docs/grafana/latest/administration/service-accounts/)
- [Grafana — Roles and permissions](https://grafana.com/docs/grafana/latest/administration/roles-and-permissions/)
- [Grafana — Team management](https://grafana.com/docs/grafana/latest/administration/team-management/)
- [Vercel — Audit Logs](https://vercel.com/docs/audit-log)
- [Vercel — Access Roles](https://vercel.com/docs/rbac/access-roles)
- [Vercel — Audit Logs with SIEM integration GA changelog](https://vercel.com/changelog/audit-logs-with-siem-integration-now-generally-available)
- [Vercel — Account management](https://vercel.com/docs/accounts)
- [Linear — Members and roles](https://linear.app/docs/members-roles)
- [Linear — Audit log](https://linear.app/docs/audit-log)
- [Linear — Audit log changelog](https://linear.app/changelog/2021-10-07-audit-log)
- [Linear — Workspaces](https://linear.app/docs/workspaces)
- [Notion — Audit log](https://www.notion.com/help/audit-log)
- [Notion — Administer your workspace](https://www.notion.com/help/category/enterprise-admin)
- [Notion — Organization-level controls](https://www.notion.com/help/organization-level-controls)
- [Notion — Enterprise security provisions](https://www.notion.com/help/guides/notion-enterprise-security-provisions)
- [Notion developers — Audit log events](https://developers.notion.com/compliance/audit-log-events)
- [GitHub Enterprise — Site admin dashboard](https://docs.github.com/en/enterprise-server@3.5/admin/configuration/configuring-your-enterprise/site-admin-dashboard)
- [GitHub Enterprise — Auditing users across your enterprise](https://docs.github.com/en/enterprise-server@3.10/admin/managing-accounts-and-repositories/managing-users-in-your-enterprise/auditing-users-across-your-enterprise)
- [Stripe blog — New roles and permissions in the Dashboard](https://stripe.com/blog/new-roles-and-permissions-in-the-dashboard)
- [Stripe — User roles](https://docs.stripe.com/get-started/account/teams/roles)
- [Stripe — Manage access to your organization](https://docs.stripe.com/get-started/account/orgs/team?locale=en-GB)
- [Stitchflow — Stripe user management guide](https://www.stitchflow.com/user-management/stripe/manual)

### Patterns and engineering guides

- [AverageDevs — Designing audit logs for SaaS](https://www.averagedevs.com/blog/audit-logs-saas-compliance-trust)
- [AuditKit — Stream audit logs to Splunk, Datadog, or Elastic](https://auditkit.dev/blog/siem-integration-audit-logs)
- [WorkOS — Developer's guide to audit logs / SIEM](https://workos.com/blog/the-developers-guide-to-audit-logs-siem)
- [Yaro Labs — Safe user impersonation tool for SaaS support](https://yaro-labs.com/blog/user-impersonation-tool-saas)
- [Yaro Labs — How to build a SaaS admin panel](https://yaro-labs.com/blog/saas-admin-panel)
- [OneUptime — Implementing presence detection with WebSockets](https://oneuptime.com/blog/post/2026-02-02-websocket-presence-detection/view)
- [Pusher — What are WebSockets](https://pusher.com/websockets/)
- [System Design One — Real-time presence platform](https://systemdesign.one/real-time-presence-platform-system-design/)
- [OWASP — Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [OnSecurity — Session management vulnerabilities](https://onsecurity.io/article/session-management-vulnerabilities-what-developers-get-wrong-and-how-to-fix-them/)
- [NocoBase — How to design an RBAC system](https://www.nocobase.com/en/blog/how-to-design-rbac-role-based-access-control-system)
- [AltexSoft — Access Control Matrix practical guide](https://www.altexsoft.com/blog/access-control-matrix-acm/)
- [Frontegg — Access control matrix: key components & 5 best practices](https://frontegg.com/blog/access-control-matrix)
- [Lumos — Access control matrix implementation guide](https://www.lumos.com/topic/access-control-matrix-implementation-guide)
- [GitNexa — SaaS dashboard UX patterns 2026](https://www.gitnexa.com/blogs/saas-dashboard-ux-patterns)
- [F1Studioz — Smart SaaS dashboard design guide](https://f1studioz.com/blog/smart-saas-dashboard-design/)
- [Setproduct — Best SaaS dashboard examples](https://www.setproduct.com/blog/saas-dashboard-examples)
- [Carlos Smith on Medium — Admin Dashboard UI/UX best practices for 2025](https://medium.com/@CarlosSmith24/admin-dashboard-ui-ux-best-practices-for-2025-8bdc6090c57d)

### Standards and accessibility

- [RFC 9457 — Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)
- [Swagger blog — Problem Details (RFC 9457): hands-on with API error handling](https://swagger.io/blog/problem-details-rfc9457-api-error-handling/)
- [Redocly — RFC 9457: Better information for bad situations](https://redocly.com/blog/problem-details-9457)
- [WebAIM — WCAG 2 Checklist](https://webaim.org/standards/wcag/checklist)
- [Deque — Web Accessibility Checklist (WCAG 2.2 AA)](https://media.dequeuniversity.com/en/docs/web-accessibility-checklist-wcag-2.2.pdf)
- [Accessible.org — WCAG Checklist 2.1 AA and 2.2 AA](https://accessible.org/wcag/)
- [Tableau — Build dashboards for accessibility](https://help.tableau.com/current/pro/desktop/en-us/accessibility_dashboards.htm)

### Design references (visual)

- [Dribbble — audit log tag](https://dribbble.com/tags/audit-log)
- [Dribbble — audit dashboard search](https://dribbble.com/search/audit-dashboard)
- [Dribbble — timeline filtering UI by Mike Hince](https://dribbble.com/shots/15449152-Timeline-filtering-UI-Design)
- [Dribbble — Dark Mode Panel by Diana Larussa](https://dribbble.com/shots/25972017-Dark-Mode-Panel-Diagram-Management-Dashboard-User-Flows-SaaS)
- [Dribbble — Dark Mode Admin Dashboard by Dmitry Sergushkin](https://dribbble.com/shots/21881540-Dark-Mode-Admin-Dashboard)
- [Dribbble — SaaS dashboard tag](https://dribbble.com/tags/saas-dashboard)
- [Dribbble — admin dashboard tag](https://dribbble.com/tags/admin-dashboard)

### MetalDocs internal anchors

- `wiki/decisions/0016-view-grade-capabilities.md` — view-grade capability registry (gates every tab in the IA)
- `wiki/architecture/api-contract.md` + `wiki/architecture/api-design-system.md` — contract-first envelopes
- `wiki/architecture/frontend-structure.md` — canonical FE layout (where the new feature lives)
- `wiki/concepts/authz-tiers.md` — two-tier authz model that the cap matrix surfaces
- `wiki/quality/screen-qa-checklist.md` + `wiki/quality/backend-api-qa-checklist.md` — gating QA checklists for shipping the new screen

