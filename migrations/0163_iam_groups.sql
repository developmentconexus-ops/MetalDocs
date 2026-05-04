CREATE TABLE IF NOT EXISTS metaldocs.iam_groups (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS metaldocs.iam_group_members (
  group_id UUID NOT NULL REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL,
  tenant_id UUID NOT NULL,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by TEXT,
  PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS metaldocs.iam_group_roles (
  group_id UUID NOT NULL REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE,
  role TEXT NOT NULL,
  PRIMARY KEY (group_id, role)
);
