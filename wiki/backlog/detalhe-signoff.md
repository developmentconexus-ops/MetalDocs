# Backlog — Detalhe Signoff (F5.1)

Deferred-with-trigger items surfaced while building the Detalhe Signoff cockpit
(`/approvals/:documentId`). Each row names the concrete unblock condition.

| Item | Why deferred | Unblock trigger | Owner |
|------|--------------|-----------------|-------|
| "Mudanças vs versão anterior" diff tab | No document-diff backend exists. `GET /documents/{id}/view` returns a single rendered PDF pointer; there is no endpoint that compares the under-review revision against the previously published one. Rendering a fabricated diff would violate the honesty rule. | A backend diff endpoint (e.g. `GET /documents/{id}/diff?against={revision}`) returning structured added/removed/changed regions or a diff artifact. When it ships, wire the tab to it. | Frontend (consume) + Backend (produce) |

Until then the tab renders an honest explanation and no invented data.
