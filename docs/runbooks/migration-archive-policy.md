# Runbook: Migration Archive Policy

Historical migrations may be classified into:

- required for upgrade history
- historical-only
- destructive/dev-reset artifacts

No file moves or deletions happen until:

1. baseline bootstrap is validated
2. legacy replay remains available
3. archive candidates are reviewed explicitly

Historical migrations remain evidence until the curated baseline and legacy replay gates pass. Archive classification must not remove the ability to debug existing DB upgrade history.
