-- Plan 6: seed audit.read capability for system_admin
INSERT INTO metaldocs.role_capabilities (role, capability, created_at)
VALUES ('system_admin', 'audit.read', NOW())
ON CONFLICT (role, capability) DO NOTHING;
