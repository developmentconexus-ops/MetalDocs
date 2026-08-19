# R10 Rebaseline Cockpit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a standalone operator-facing HTML cockpit that visually projects the current MetalDocs R10 rebaseline state, decisions, implementation-readiness path, authorities, T8-A workbench, and a copyable fresh-session prompt without becoming architecture authority.

**Architecture:** One static file at `docs/operator/r10-rebaseline-cockpit.html` contains all markup, CSS and vanilla JavaScript. It has no external dependencies and reads no live state; every status claim is a dated projection of current durable authority, with permanent authority disclaimers and source paths.

**Tech Stack:** HTML5, CSS3, vanilla JavaScript, temporary Python/BeautifulSoup structural validation, Chromium/Playwright interaction and responsive checks.

**Spec:** `docs/superpowers/specs/2026-08-19-r10-rebaseline-cockpit-design.md`

## Global Constraints

- `VISUAL ORIENTATION / PROJECTION — NOT ARCHITECTURE AUTHORITY` must be visible without interaction.
- Router authority is `wiki/architecture/r10-technical-architecture.md`.
- Current projection: T1→T7 CLOSED; T8-A ACTIVE; T8-B→T12 NOT OPEN; implementation BLOCKED.
- No framework, package, build step, CDN, remote asset, live API/GitHub call or generator.
- No numeric implementation-readiness percentage or ETA.
- No product implementation code.
- Last synchronized date is 2026-08-19; fresh-session prompt must revalidate current PR/HEAD.

---

### Task 1: Create cockpit semantic structure and visual system

**Files:**
- Create: `docs/operator/r10-rebaseline-cockpit.html`

**Interfaces:**
- Consumes: current R10 router, T1→T7 durable authorities, post-T6 program, TRRB, active T8-A bootstrap.
- Produces: stable section IDs used by nav/search/filter/interaction checks: `overview`, `path`, `architecture`, `scope`, `decisions`, `guardrails`, `t8a`, `open-decisions`, `authorities`, `journey`, `session`.

- [ ] **Step 1: Write a temporary structural contract check before the HTML exists**

Create `/tmp/check_r10_cockpit.py` that fails unless the target contains the required title, authority disclaimer, all eleven section IDs, current stage strings, T7 no-migration decision, no external `http(s)` asset references, and buttons with IDs `copy-session-prompt`, `copy-authority-chain`, `print-cockpit`.

- [ ] **Step 2: Run structural check and verify RED**

Run: `python /tmp/check_r10_cockpit.py /mnt/data/r10-rebaseline-cockpit.html`

Expected: non-zero exit because the cockpit file does not exist.

- [ ] **Step 3: Implement the minimal complete standalone HTML**

Create the full document with inline CSS/JS, sticky navigation, executive state, stage pipeline, semantic architecture map, scope tiers, decision cards, guardrails, T8-A workbench/matrix, scheduled open decisions, authority navigator, journey timeline and session prompt.

- [ ] **Step 4: Run structural check and verify GREEN**

Run the same command; expected exit 0 and a summary of required contracts found.

### Task 2: Validate browser interactions and responsive behavior

**Files:**
- Modify if needed: `docs/operator/r10-rebaseline-cockpit.html`

**Interfaces:**
- Consumes: element IDs/data attributes from Task 1.
- Produces: working search/filter/accordion/copy/print behavior and responsive layout.

- [ ] **Step 1: Write a temporary Playwright browser check**

Create `/tmp/check_r10_cockpit_browser.py` that launches `/usr/bin/chromium` headless against the local file and asserts: no page errors; nav anchors exist; search for `REV000` leaves matching content visible; Active filter leaves `T8-A` visible; first decision accordion toggles; copy-session-prompt shows success feedback; desktop width has sidebar nav; 390px mobile width does not overflow horizontally.

- [ ] **Step 2: Run browser check before interaction fixes**

Expected: any missing interaction/layout behavior fails with a precise assertion.

- [ ] **Step 3: Make the smallest HTML/CSS/JS corrections needed**

No new features beyond the approved design.

- [ ] **Step 4: Re-run browser check until GREEN**

Expected: exit 0, zero JS/page errors, all interaction/responsive assertions pass.

- [ ] **Step 5: Capture desktop and mobile screenshots for visual inspection**

Create temporary screenshots under `/tmp/`; inspect for clipping, illegible hierarchy, overlapping sticky elements, accidental horizontal scroll and misleading status emphasis.

### Task 3: Publish, route and verify the operator artifact

**Files:**
- Create in repo: `docs/operator/r10-rebaseline-cockpit.html`
- Modify: `README.md`
- Remove after completion: `docs/superpowers/specs/2026-08-19-r10-rebaseline-cockpit-design.md`
- Remove after completion: `docs/superpowers/plans/2026-08-19-r10-rebaseline-cockpit.md`

**Interfaces:**
- Consumes: validated local cockpit bytes.
- Produces: durable non-authoritative cockpit discoverable from root README; staging preserved only in Git history.

- [ ] **Step 1: Upload exact validated HTML bytes to the scoped PR branch**

Use GitHub contents write on `docs/a8-authz-approval-redesign-ledger`.

- [ ] **Step 2: Add a small root README orientation entry**

Add `docs/operator/r10-rebaseline-cockpit.html` with explicit `non-authoritative visual orientation` wording and route readers back to `wiki/architecture/r10-technical-architecture.md`.

- [ ] **Step 3: Fresh-fetch cockpit and README from GitHub**

Verify cockpit title/status/disclaimer and README link from the branch, not from local assumptions.

- [ ] **Step 4: Delete completed spec/plan staging**

Per repository documentation governance, Git history remains provenance; live `docs/superpowers` stays active-stage staging only.

- [ ] **Step 5: Fresh-revalidate PR/HEAD and report verification limits**

Confirm PR open/draft/mergeable and current HEAD. Report browser/structural verification separately from repository-wide `tools/verify --profile=pr`, which is not proven unless actually run in a checkout.
