# F6.6 Plan

1. Verify Node.js/npm available (`node --version`, `npm --version`).
2. Run `npm run gen:api` from `frontend/apps/web/`.
3. Run `git diff --stat frontend/apps/web/src/lib/api-types/index.d.ts` — confirm 8 new types.
4. Run H-D exit-bar grep — confirm 0 matches.
5. Run `go build ./...` — must be clean.
6. Run `go test -count=1 ./...` — all ok.
7. Edit `wiki/architecture/api-contract.md` Last-verified line to 2026-06-19 M6 stamp.
8. Author F6.6 spec/plan/evidence artifacts.
9. Commit: `feat(m6/f6.6): FE codegen regen + wiki stamp + H-D Grep A = 0 proof`.
