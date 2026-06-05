ALTER TABLE metaldocs.iam_user_roles
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid;
