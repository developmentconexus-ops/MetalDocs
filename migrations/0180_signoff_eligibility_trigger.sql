-- 0180_signoff_eligibility_trigger.sql
-- Defense in depth for J1: rejects INSERT into approval_signoffs if actor
-- is not in the parent stage's eligible_actor_ids jsonb snapshot. Belt to
-- the application-layer CheckEligibility braces.

BEGIN;

CREATE OR REPLACE FUNCTION enforce_signoff_eligibility() RETURNS trigger AS $$
DECLARE
    eligible jsonb;
BEGIN
    SELECT eligible_actor_ids INTO eligible
    FROM   approval_stage_instances
    WHERE  id = NEW.stage_instance_id;

    IF eligible IS NULL OR NOT eligible @> to_jsonb(NEW.actor_user_id::text) THEN
        RAISE EXCEPTION 'signoff: actor % is not in eligible_actor_ids for stage %',
            NEW.actor_user_id, NEW.stage_instance_id
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql
   SET search_path = pg_catalog, pg_temp;

DROP TRIGGER IF EXISTS enforce_signoff_eligibility_trg ON approval_signoffs;
CREATE TRIGGER enforce_signoff_eligibility_trg
BEFORE INSERT ON approval_signoffs
FOR EACH ROW EXECUTE FUNCTION enforce_signoff_eligibility();

INSERT INTO public.schema_migrations (version, description)
VALUES ('0180', 'enforce signoff eligibility against eligible_actor_ids snapshot')
ON CONFLICT (version) DO NOTHING;

COMMIT;
