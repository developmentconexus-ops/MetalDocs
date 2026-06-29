-- SP-2: dictionary token substitution introduces a fourth placeholder-value
-- source, 'dictionary', pinned at document creation (source of truth = the
-- tenant token dictionary; the document captures a pinned copy per revision).
-- The source CHECK is closed at baseline ('user','computed','default'); widen it
-- to include 'dictionary'. Additive — preserves the existing three. Forward-only,
-- idempotent (DROP IF EXISTS then ADD).

BEGIN;

ALTER TABLE public.document_placeholder_values
    DROP CONSTRAINT IF EXISTS document_placeholder_values_source_check;

ALTER TABLE public.document_placeholder_values
    ADD CONSTRAINT document_placeholder_values_source_check
    CHECK (source = ANY (ARRAY['user'::text, 'computed'::text, 'default'::text, 'dictionary'::text]));

INSERT INTO public.schema_migrations (version, description)
VALUES ('0249', 'widen document_placeholder_values source CHECK to add dictionary (SP-2 dictionary token substitution)')
ON CONFLICT (version) DO NOTHING;

COMMIT;
