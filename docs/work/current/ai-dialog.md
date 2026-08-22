# T10 Fable independent review — Round 2

> **Evidence only — non-authoritative. This review branch must never merge.**

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T10 — Transition / Cutover
Candidate branch          arch/t10-transition-cutover
Exact corrected HEAD      c1afc292bc94f48bfd2146c3b4374342ff5c2701
Required candidate CI     #1157 SUCCESS
Candidate Draft PR        #158
Round-1 Evidence PR       #159 CLOSED / UNMERGED
Round-1 verdict           NOT CONVERGED / MATERIAL=3
Review branch             review/t10-fable-r2
Round                     2 / BOUNDED CONFIRMATION
```

## Scope

This is **not** a fresh unconstrained redesign. Confirm the Round-1 corrections and detect regressions only.

Read strictly:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t10-transition-cutover.md
→ only exact owner authority needed to verify a challenged correction
```

## Confirm F1 closure — B2 clean seal / B3 authority edge

Verify that:

```text
B2 proof binds to the exact deployed production candidate/profile
proof-fixture Product truth is removed before authority begins
a mechanical clean-baseline verification closes B2
clean-seal evidence is operations/provenance only, never Product state
affected B2 evidence is re-armed after reset/rebuild
proof mutation paths are fenced after the clean seal
any unexpected post-seal Product mutation blocks destructive reset pending classification
B3 = first post-seal authoritative Product bootstrap commit
no Product activation marker/table/endpoint/Permission/owner is invented
```

Attack both old failure directions:
1. proof fixture silently promoted into authority;
2. committed authority destroyed by a supposedly pre-B3 reset.

The Lead deliberately **did not** adopt a Product activation marker. Challenge whether the corrected clean-seal + first-post-seal Product commit law is mechanically decidable without one.

## Confirm F2 closure — authoritative recovery point before B4

Verify that B4 cannot expose ordinary serving unless:

```text
B3 authoritative baseline exists
target remains ready
at least one complete authoritative R10 recovery point covers the current B3 baseline
complete-set/manifest/exact-content integrity checks pass
```

T8-G restore capability / isolated restore drill belongs B2. T10 only sequences an actual authoritative recovery point after B3 and before B4.

Also challenge the Lead refinement:

```text
loss of canonical R10 authority + every coherent authoritative recovery point
→ catastrophic authority loss
→ remain fail-closed
→ no automatic re-bootstrap / disposable-state promotion
→ explicit operator/business recovery decision + smallest architecture reopen required
```

Confirm this is more truthful than treating re-bootstrap as ordinary recovery.

## Confirm F3 closure — serving fence / cleanup

Verify that B4 requires every inventoried user-reachable disposable DEV/test serving path to be unable to accept ordinary Product mutations, including stale DNS/cache/direct-origin/bookmark routes.

DNS switch alone must not count as fencing.

Verify cleanup now requires:

```text
contains no business truth requiring bounded reopen
```

with no temporal `pre-R10` loophole. Any unexpected business truth in a supposedly disposable estate must stop cleanup.

## Confirm bounded MINOR handling

- F4: exact production candidate/profile wording; reset-dependent proof re-armed.
- F5: T3 non-serving bootstrap concern is the semantic anchor. **Do not** reinterpret T8-D `bootstrap/provisioner` trust class as Product bootstrap; it is provisioning/DDL-only. If later implementation proves accepted T8-G surfaces cannot realize semantic bootstrap, only then is bounded T8-G reopen triggered.
- F6: external estate is unclassified until B0 proof; expected DEV/test status is never disposal permission.

## Regression envelope

Must remain:

```text
barriers                         exactly B0→B4 / 5 total
historical business migration   absent
business authority              singular
application operations          78
operation 79                    absent
new Permission                  none
new semantic owner              none
new Product state               none
new runtime capability          none in current T10 candidate
T11/T12                         not open
Product implementation          blocked
```

A possible future T8-G reopen named as a conditional implementation falsifier is not a current runtime-capability addition.

## Finding standard

Classify findings:

```text
MATERIAL
MINOR
NOTE
```

A new MATERIAL finding must identify:
1. exact corrected candidate claim;
2. exact accepted authority/property violated or still uncovered;
3. reachable failure sequence;
4. smallest correction/upstream reopen;
5. why Round 1 did not already cover it.

## Required verdict

End exactly:

```text
VERDICT = CONVERGED | NOT CONVERGED
MATERIAL findings = N
Round 3 justified = YES | NO
```

Write review Evidence below this line only.

---

## Fable Round 2 response
