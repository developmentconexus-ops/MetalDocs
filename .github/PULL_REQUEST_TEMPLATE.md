## Summary

## Type of change
- [ ] API
- [ ] Domain
- [ ] Infra
- [ ] UI
- [ ] Docs only

## Governance checklist
- [ ] I read and followed `AGENTS.md`
- [ ] OpenAPI updated (if API changed)
- [ ] ADR/RFC linked when required
- [ ] Tests added/updated per standards
- [ ] Runbook updated (if infra/ops changed)

## Security review (REQ-SEC-3 — `wiki/quality/security-review-checklist.md`)
Does this diff touch **auth, input handling, file paths, crypto, or queries**?
See the checklist's "When a change is in scope" table for the precise trigger definition.
- [ ] N/A — none of the five trigger areas apply
- [ ] Applies — trigger area(s): _____________
  - [ ] relevant ASVS chapter(s) checked against this diff: _____________
  - [ ] any Level-1 deviation is fixed or named as an explicit, linked exception
  - [ ] tests exist for the security-relevant behavior, not only the happy path

> This block is a reviewer prompt, not a CI gate — see the checklist's "Global maximum this defers" section for the mechanically-enforced version and the milestone that ships it.

## Evidence
- [ ] Unit test output
- [ ] Contract/Integration/E2E output (as applicable)
- [ ] Screenshots/logs (if applicable)

## Risks and rollback
