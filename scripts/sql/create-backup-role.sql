-- Execute como usuario admin/superuser no PostgreSQL.
-- Ajuste senha e database conforme ambiente.

-- NOTA (M7 RLS-truth sweep): varias tabelas (ex.: metaldocs.audit_events,
-- metaldocs.approval_signoffs) sao FORCE ROW LEVEL SECURITY. Um role apenas
-- com SELECT, sem o atributo BYPASSRLS, falha um pg_dump full-DB nessas
-- tabelas com "query would be affected by row-level security policy". Por
-- isso o role recebe BYPASSRLS abaixo: isso permite ler alem da RLS sem
-- conceder superuser — o role continua nao-superuser e somente-leitura
-- (SELECT), apenas ganha a capacidade de COPY sobre tabelas FORCE RLS.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_backup') THEN
    CREATE ROLE metaldocs_backup LOGIN PASSWORD 'CHANGE_ME_BACKUP_PASSWORD' BYPASSRLS;
  END IF;
END$$;

-- Idempotente: garante BYPASSRLS mesmo se o role ja existia antes desta
-- alteracao (roles criados pela versao anterior deste script).
ALTER ROLE metaldocs_backup BYPASSRLS;

GRANT CONNECT ON DATABASE metaldocs TO metaldocs_backup;

-- Schema metaldocs (dados de negocio).
GRANT USAGE ON SCHEMA metaldocs TO metaldocs_backup;
GRANT SELECT ON ALL TABLES IN SCHEMA metaldocs TO metaldocs_backup;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA metaldocs TO metaldocs_backup;

ALTER DEFAULT PRIVILEGES IN SCHEMA metaldocs
  GRANT SELECT ON TABLES TO metaldocs_backup;

ALTER DEFAULT PRIVILEGES IN SCHEMA metaldocs
  GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_backup;

-- Schema public (River jobs + migrations, ~30 tabelas). O backup e full-DB
-- (pg_dump --format custom sem escopo de schema), entao o role precisa
-- tambem enxergar o schema public, nao apenas metaldocs.
GRANT USAGE ON SCHEMA public TO metaldocs_backup;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO metaldocs_backup;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO metaldocs_backup;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO metaldocs_backup;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO metaldocs_backup;

