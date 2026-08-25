# T11 — B11 Access Administration — clean P8 rebaseline

> **Status:** R3 LOCKED / OPERATOR-APPROVED.  
> **Block:** B11 — Access Administration.  
> **Route:** `/admin/access`.  
> **Branch baseline:** `origin/main` at `58b55e0f518cd8652e08a1b1fa79fb86e7beb218`.  
> **Implementation:** BLOCKED.  
> **Artifact class:** temporary frontend-planning Evidence under `docs/work/**`; it must not enter the eventual merge candidate/main.

## 1. Current authority pack

```text
AGENTS.md
→ docs/roadmap.md
→ Engineering Method v1.0.0
→ Frontend Product Experience Planning Method v2.3
→ docs/index.md
→ docs/decisions/access-assignment-read.md
→ relevant Authorization / Wire / Frontend sections
```

Challenge R1 preserved the four clean-rebaseline closures but falsified the first locked bytes in two material local respects: Role bundles and non-operable 403/404 Evidence. R2 corrects that smallest scope plus the accepted drawer-accessibility mismatch. See `t11-b11-final-challenge-r1.md`.

Current semantic constraints preserved:

```text
RoleAssignment = Subject(User | Group) × fixed Role × Scope(Company | Area)
GroupMembership mutation = access.manage
RoleAssignment mutation = access.manage
Group identity remains Organization-owned
Group.area_id does not exist
frontend does not compute effective access
Role/Permission vocabulary remains fixed and read-only
89 application operations remain unchanged
```

## 2. Evidence use

PR #173 and its R5/R6/R7 artifacts were used only as historical Evidence.

Preserved where current authority still supports them:

```text
/admin/access
Por Área / Grupos / Funções
Area-specific and Company-wide grants in separate regions
Group multi-scope footprint
fixed Role meaning
Subject × Role × Scope grant review
exact assignment_id revoke
bounded membership consequence copy
visible pagination and continuation-failure semantics
```

Not inherited:

```text
old B11 LOCK state
complete GroupMembership knowledge in the add-member picker
pre-pagination User eligibility filtering
duplicate fixture grant on repeated same-key confirmation
PR #173 workspace chronology or status
```

The current candidate is a new integrated reconstruction from `main`, not a continuation of the superseded branch.

## 3. Canonical P8 candidate

Entry:

```text
docs/work/current/t11-b11-access-administration-p8.html
```

Supporting assets:

```text
docs/work/current/t11-b11-access-administration-p8.css
docs/work/current/t11-b11-access-administration-p8.js
```

The three files form one browser-operable low-fidelity P8 artifact. Visual styling is deliberately disposable; the candidate tests interaction structure and server-truth boundaries only.

## 4. Four known failure classes closed in the candidate

### F1 — hidden all-page traversal

Every potentially unbounded supporting or canonical read used by B11 exposes visible Previous/Next traversal:

```text
op6  UserPage
op16 AreaPage
op22 GroupPage
op27 GroupMemberPage
op31 filtered RoleAssignmentPage
```

The UI presents page number plus `há mais` / `fim da lista`, never a total count. A deterministic continuation failure preserves the current page and explicitly refuses to call it complete.

### F2 — raw op6 page fidelity

Both User pickers consume the same raw fixture boundaries:

```text
page 1  João / Beatriz / Rafael / Ana
page 2  Bruno / Carla / Paulo DISABLED / Sofia
page 3  Luciana / Diego / Mariana / Felipe
```

No state filter runs before pagination. Paulo remains visible on page 2 and cannot be selected.

### F3 — incomplete membership knowledge

The add-member picker never disables a User merely because the browser believes the relation exists or not.

```text
known from already loaded op27 page
  → optional local guidance only

not seen in loaded op27 pages
  → remains unknown

op28 PUT
  → 201 first relation
  → 204 relation already existed
```

`Sofia Barros` is deliberately already in the selected Group but lives on a later op27 page. Selecting her from raw op6 page 2 produces `204` and zero new membership mutation. `Mariana Costa` on op6 page 3 produces `201`.

### F4 — repeated same-key grant

The grant simulation stores the first successful result by:

```text
Idempotency-Key + normalized command fingerprint
```

Same key + same fingerprint replays the same `assignment_id` and leaves the semantic mutation counter unchanged. The fixture supports both:

```text
ordinary completed replay
ambiguous transport after server commit → same-key retry
```

After success, the Product confirmation is terminal in the current dialog. A `REVIEW ONLY` proof control can replay the key without exposing a second Product mutation action.

## 5. B11-F1 probes retained

The P8 can still falsify the ratified op31 precision:

1. `Aprovadores Financeiro` has Roles in multiple Areas plus Company scope.
2. The Group lens shows that footprint through filtered op31 pages.
3. The Area lens keeps exact Area and Company-wide assignments in separately labeled regions.
4. Role meaning comes only from fixed server-returned RoleView fixtures and is read-only.
5. Revocation targets one exact `assignment_id`.
6. Active op31 filters remain visibly paginated.
7. No surface claims complete per-User effective access.
8. No `Group.area_id` or equivalent ownership appears.

## 6. Suggested operator walkthrough

### A — Area and Group comprehension

1. In `Por Área`, inspect `COM · Comercial`.
2. Confirm that Area-specific and Company-wide sections remain separate.
3. Open `Grupos` and inspect `Aprovadores Financeiro`.
4. Traverse its access pages and confirm multiple Area scopes plus Company scope remain independent grants.
5. Confirm the Group never appears to belong to one Area.

### B — real page traversal

1. Traverse op22 Groups, op27 Members and an op31 grant list.
2. Arm `Próxima página → falha` before a continuation.
3. Confirm the already loaded page remains visible and is not labeled complete.
4. In Members, reach a later page and remove one exact User.

### C — add-member reconciliation

1. Reset the fixture and open `Grupos → Aprovadores Financeiro → Adicionar membro`.
2. Go to raw op6 page 2.
3. Confirm `Paulo Mendes — DISABLED` remains visible and unavailable.
4. Select `Sofia Barros` without traversing op27 beyond its first page.
5. Confirm membership and observe `204`, with mutation count unchanged.
6. Reset; return to the picker, traverse to page 3, select `Mariana Costa` and observe `201`.

### D — grant idempotency

1. Reset and open `Conceder acesso`.
2. Traverse raw User pages; verify page boundaries and DISABLED behavior.
3. Select `Mariana Costa`, choose a Role and Scope, then review.
4. Confirm once and record the `assignment_id`, key and grant mutation count.
5. Use `Provar replay da mesma chave` and verify identity/count remain unchanged.
6. Reset; arm `Próximo grant → ambíguo após commit`.
7. Compose and confirm a grant, then use `Repetir o mesmo comando`.
8. Verify the retry returns the committed identity with zero second mutation.

### E — structural checks

1. Navigate the three local tabs by keyboard arrows.
2. At a narrow viewport, open the global drawer and inspect stacked panels/dialogs.
3. Confirm visible focus, labeled controls and non-color-only state text.
4. Confirm no custom Role editor, effective-access matrix, hidden search or operation 90 appears.

## 7. Gate

```text
P8 implementation candidate       COMPLETE LOCALLY
R2 implementation candidate        COMPLETE LOCALLY
R2 static/browser verification     PASS
prior operator LOCK                SUPERSEDED BY MATERIAL EVIDENCE
R2 operator walkthrough            APPROVED
R2 operator LOCK                   SUPERSEDED BY MATERIAL R2 CHALLENGE
R3 implementation candidate        COMPLETE LOCALLY
R3 operator walkthrough            APPROVED
R3 operator LOCK                   LOCKED / 2026-08-25
P9 Screen Contract                 COMPLETE / PASS AGAINST R3 LOCK
P10 pattern pass                   COMPLETE / PASS AGAINST R3 LOCK
independent adversarial gate       CONVERGED WITH NON-BLOCKING NOTE
```

Post-LOCK proofs:

```text
docs/work/current/t11-b11-screen-contract.md
docs/work/current/t11-b11-pattern-consolidation.md
```

Verification below records the superseded first-lock bytes and must not be read as R2 closure:

```text
node --check                       PASS
git diff --check                   PASS
desktop DOM/render                 PASS
browser console errors/warnings    0
visible continuation failure       PASS / page preserved
raw op6 page 2 + DISABLED User     PASS
op28 existing relation             204 / mutations 0
op28 first relation                201 / mutations 1
completed grant replay             same id/key / mutations 1 → 1
ambiguous post-commit retry        same id/key / mutations 1 → 1
later op27 removal                 PASS
keyboard tab arrows                PASS
390×844 drawer/stack/dialog fit    PASS
```

### R2 candidate verification after challenge R1

Pre-status candidate identity operated by the user:

```text
HTML  02fb202ff8031fc4f2a760537d53cb4680eb515b
CSS   9ce012007613777187ae70956c2bfa09e7066c16
JS    3e923746cfea01142249b8166500833c807c5ce5
```

Exact post-decision locked package:

```text
HTML  c0dfe7b942b83f53374307dbdb3d3524b7d47c69
CSS   9ce012007613777187ae70956c2bfa09e7066c16
JS    3e923746cfea01142249b8166500833c807c5ce5
LOCK  docs/work/current/t11-b11-p8-r2-operator-relock.md
```

Browser/static proof:

```text
node --check                                      PASS
HTTP / current R2 title/status                    200 / PASS
six exact Role bundles/scopes                     PASS
/admin/access 403 replaces surface/actions        PASS
selected Group 404 removes identity/detail/action PASS
404 select-next-disclosed-Group recovery          PASS
drawer aria-controls/focus/inert/Escape/return    PASS
raw op6 page 2 + DISABLED User                    PASS
op28 existing relation                            204 / mutations 0
op28 first relation                               201 / mutations 1
completed grant replay                            same id/key / mutations 1 → 1
ambiguous post-commit retry                       same id/key / mutations 1 → 1
browser console                                   0
```

### R3 candidate verification after challenge R2

Candidate identity — not LOCKED until the operator approves these exact bytes:

```text
HTML  9642cce8b8a45ade8005fcf299a8ca69ff8d5921
CSS   9ce012007613777187ae70956c2bfa09e7066c16
JS    670ff9b905d94014ff27698e2a23c868316030a4
```

Exact post-decision R3 locked package:

```text
HTML  ea20912e5259f4f3f51df7ce09ee3f2e5cfc7540
CSS   9ce012007613777187ae70956c2bfa09e7066c16
JS    670ff9b905d94014ff27698e2a23c868316030a4
LOCK  docs/work/current/t11-b11-p8-r3-operator-relock.md
```

Focused temporal proof:

```text
ambiguous committed command key                  K1 / 00000000-0000-4000-8000-000000000001
Subject / Role / Scope after ambiguity           disabled / disabled / disabled
Revisar / Fechar after ambiguity                 disabled / disabled
same-key retry                                   enabled / only resolution
Escape                                            dialog remains open
retry result                                      asg-100 / same K1
semantic grant mutations                          1 → 1
close after resolved replay                       enabled
browser console                                   0
```

## 8. Prior operator disposition and reopen

The operator viewed the current P8 HTML in the in-app browser and replied `Aprovado` on 2026-08-25 in direct response to the explicit question whether to declare the B11 P8 LOCK or request another revision.

```text
decision owner       operator
decision             LOCK
candidate            current integrated clean P8
assistant decision   none
```

The original LOCK froze its exact interaction bytes, but challenge R1 later falsified that candidate. The operator separately operated and approved R2; `t11-b11-p8-r2-operator-relock.md` owns that superseded re-LOCK. Challenge R2 then falsified only the alternate op32 ambiguity path. The operator separately operated and approved R3; `t11-b11-p8-r3-operator-relock.md` owns the current exact LOCK. None of these decisions authorizes Product implementation, P11, B12 or any Product/backend capability excluded by current authority.
