# Database Overview

MetalDocs uses PostgreSQL as the source of truth for product state.

The database is governed by:

- curated baseline for fresh environments
- product reference data for required roles/capabilities/system records
- optional local dev seeds for developer workflows
- post-baseline forward migrations for new changes
- legacy historical migrations for recovery/debugging until explicit archive approval

Runtime truth comes first. A database object belongs in the curated baseline only when current runtime code, product behavior, or an explicit documented debt note justifies it.
