# AuthZ grant-model unification — operator decisions (2026-08-06)

**Status:** decisions ratified by the operator; design not started. This is the input record for a
future program, NOT an ADR. The ADR is written inside that program, after the
`developing-new-work` gate.

**Origin.** Surfaced by the http-surface-protocol program's Task 17 conformance suite
(`TestNoDeclaredOperationIsUnreachable`, red by design at commit `5bf81a57`). The suite reported
capabilities that no role grants on tier-1's grant path. Investigation showed the cause is not
seed data.

## The finding

Tier-1 and tier-2 read **disjoint grant tables**.

| tier | function | reads | role CHECK |
|---|---|---|---|
| 1 — HTTP edge | `CapabilityService.CanDo` (`internal/modules/iam/application/capability_service.go:48`) | `iam_user_roles` ∪ `iam_group_members`⋈`iam_group_roles` | 5 roles / **none at all** |
| 2 — in-tx | `authz.Require` (`internal/modules/iam/authz/authz.go:145`) | `role_capabilities` ⋈ `user_process_areas` | 7 roles, incl. `area_admin`, `qms_admin`, `signer` |

Three assignment tables, three different role vocabularies, hand-synced. `iam_group_roles.role`
carries no CHECK constraint at all (`db/baseline/0001_current_schema.sql:1245`).

Consequence: a principal holding only an area membership is refused at tier-1 before tier-2 ever
runs. The area model is unreachable through HTTP except to further narrow someone tier-1 already
admitted.

**This is a documented unfinished migration, not a design.** ADR 0007's own Context section
(`wiki/decisions/0007-two-tier-authz.md:22`) records it verbatim:

> "The 2026-05-02 IAM unification plan attempted to consolidate both. It unified the middleware path
> (`StaticAuthorizer` → `CapabilityService`) but left `authz.Require` reading `user_process_areas`.
> This produced apparent dual systems and confused engineers."

The Decision section then rules the two services are "distinct tiers ... **not a unification gap**."
That is an incomplete migration ratified retroactively as architecture — an unlabelled local
maximum under CLAUDE.md's Global Maximum rule. "Confused engineers" is a symptom, not a
justification.

It is also the **same defect class the http-surface-protocol program exists to remove**: five
hand-synced enumerations of route truth collapsed into one declaration. Here it is three hand-synced
enumerations of grant truth.

## What is NOT wrong

Two enforcement points are correct and stay. Edge rejects early and cheap; in-tx is the binding
decision; the DB tripwire is the last line. That is defense in depth.

Defense in depth means the **same question** verified at several points. The defect is two
**different questions** against **different tables**, free to disagree.

`role_capabilities` (role → capability bundle) is also correct and stays. It already is the
"gerente com as capabilities de gerente" model. The defect is purely in the assignment side.

## Industry evidence

The closest match to MetalDocs' shape is **Kubernetes RBAC**, which faced the same problem: roles
that sometimes apply globally, sometimes only within a scope.

- `Role` (namespace-scoped) / `ClusterRole` (cluster-wide) — the permission bundle
- `RoleBinding` / `ClusterRoleBinding` — the assignment
- Subjects: `User`, `Group`, `ServiceAccount` — one subject abstraction
- Evaluation: union of all bindings, deny by default

The load-bearing idea: **scope lives on the binding, not in a separate table.** A `RoleBinding` may
reference a `ClusterRole`, granting that bundle inside one namespace — i.e. "qms_admin's
capabilities, but only in area QA-01".

Corroborating:
- **NIST RBAC** (ANSI INCITS 359-2004): Users → Roles → Permissions. Roles are bundles; core RBAC
  has no direct permission assignment.
- **AWS IAM**: policies attached to user/group/role, scoped by ARN + Condition — scope on the
  attachment, not a parallel table.
- **Zanzibar / OpenFGA / SpiceDB**: relation tuples `object#relation@subject`. Strictly more
  powerful; justified only if per-document sharing becomes a requirement. Not today's need.

None of them has two grant tables read by two evaluators.

## Ratified decisions

### D1 — One assignment relation, scope on the binding

Collapse `iam_user_roles`, `user_process_areas`, `iam_group_roles` + `iam_group_members` into a
single binding relation carrying `(subject_kind, subject_id, role_code, scope_kind, scope_ref)` plus
the existing effective-interval and grant/revoke provenance columns. One role catalog, referenced by
FK rather than repeated as a CHECK in three places — so adding a role becomes data, not a migration.

The two tiers become two **predicates over one source**:
- tier-1: the capability is granted at **any** scope → cheap edge reject
- tier-2: the capability is granted at `tenant` **or** at the resource's area

This yields the invariant that is missing today and is the root cause of the finding:
**tier-1 must be a strict relaxation of tier-2.** If tier-1 refuses what tier-2 would grant, the
operation is unreachable — exactly what `TestNoDeclaredOperationIsUnreachable` reports. With one
source, that becomes a testable property instead of an accident.

### D2 — No bypass. `system_admin` holds every capability as ordinary grants

**Operator decision, 2026-08-06.** Today `system_admin` is a special branch in BOTH tiers
(`capability_service.go:56-68`, `authz.go` bypass path) with its own audit record (`recordBypass`).

Under the unified model it is a role like any other, bound at `scope=tenant`, whose bundle is the
full capability set. Rationale:

- A bypass is a **second regime** — the exact structure this program family exists to remove. Two
  ways to be authorized means two things to keep correct.
- Audit becomes uniform. A bypass needs its own audit path precisely because it is not a grant;
  make it a grant and the ordinary path already records it.
- "Who can do X?" becomes answerable by one query over one table, with no special case to remember.
- The capability registry already enumerates every capability, so "all of them" is derivable, not
  hand-listed.

### D3 — No direct per-user capability grants

**Operator asked this be recorded with its reasoning.** The operator's original model was
"user gets groups, OR gets separate permissions". The first half is adopted; the second half is
deliberately rejected. Kubernetes has no direct permission assignment either, for the same reasons:

1. **Auditability collapses.** With roles only, "who holds `document.publish`?" is one join. Add
   direct grants and the answer becomes "everyone with a granting role, PLUS an unbounded set of
   one-off rows" — the question stops having a cheap, complete answer. In a regulated eQMS that
   question is not a convenience; it is evidence.
2. **Review has no unit.** A role is a reviewable object: someone can look at "gerente" and judge
   whether that bundle is right. A thousand ad-hoc grants have no object to review — each is its own
   snowflake, and drift is invisible because there is nothing to drift *from*.
3. **Revocation stops being reasonable.** Removing a capability from a role revokes it everywhere,
   once. With direct grants, revocation is a search-and-destroy across rows nobody enumerated.
4. **It reintroduces exactly the defect being removed.** A direct grant is a second grant regime
   living beside the role regime — a smaller instance of the tier-1/tier-2 split.
5. **The pressure it relieves is better relieved by cheap roles.** The real need behind "just give
   this one person this one permission" is that creating a role feels expensive. D1 already fixes
   that: the role catalog is a table, so a new role is an INSERT, not a migration. Make the right
   path cheap instead of adding a wrong path.

Recorded trade-off, honestly: this costs flexibility in the exceptional case. The accepted answer
to a genuine one-off is **create a narrow role and bind it**, which keeps the audit story intact.

### D4 — Groups and areas are orthogonal; both survive

Group = **who** (a subject). Area = **where** (a scope). They look redundant today only because each
lives in its own grant table. Under D1 they are different columns of one binding, and a
capability the current schema cannot express becomes expressible: **binding a group to an area**.

## Scope note

This is a program, not a task. It touches the two-tier PDP, the DB tripwire, ADRs 0007 / 0022 /
0026, a data migration across three tables, and the user-management frontend. Per CLAUDE.md it must
pass the `developing-new-work` gate (written system-impact analysis + Green/Yellow/Red verdict)
before any design work starts.

**Deliberately deferred out of the http-surface-protocol branch.** Seeding
`distribution.read` / `notification.read` onto `iam_user_roles` there would have written the
doomed model deeper while hiding the finding. `TestNoDeclaredOperationIsUnreachable` stays red as the
evidence that justifies this program.

## The red CI lane is a ratified choice, not an oversight

**Operator decision, 2026-08-06.** Task 18 added `./apps/...` to the `test-full.yml` and
`test-nightly.yml` integration lanes, so those lanes are RED until this program lands. The operator
was offered the alternative — convert the finding to a baseline assertion ("the unreachable set is
exactly this recorded set"), which would be green today, red on any regression, and self-deleting —
and **chose to leave the lane red**, on the grounds that this program follows immediately and there
is no team for a red lane to demoralize.

Recorded so a future session does not "repair" it:
- Do NOT skip, exclude, or baseline `TestNoDeclaredOperationIsUnreachable`.
- Do NOT remove `./apps/...` from the integration lanes to restore green.
- The accepted cost is that while the lane is red it has **no regression-detection power** — a sixth
  dead capability would not be noticed, because the lane is already failing. That is understood and
  accepted for the short window before this program starts.
- The lane goes green when this program makes the unreachable set empty. That is the fix.
