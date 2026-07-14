# Database Schemas

## `metaldocs`

Preferred schema for application-owned business and platform objects.

## `public`

Contains legacy and current objects created by historical unqualified migrations. New objects should not be added here unless a migration plan explicitly requires it.

Every retained `public` object must be marked in its table dictionary page as:

- intentionally public for now
- candidate for future move
- historical/unused candidate for exclusion
