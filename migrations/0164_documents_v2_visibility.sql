ALTER TABLE public.documents_v2
  ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'area'
  CONSTRAINT documents_v2_visibility_check
  CHECK (visibility IN ('public', 'area', 'restricted'));
