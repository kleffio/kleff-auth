DROP TRIGGER IF EXISTS users_set_timestamp ON users;
DROP INDEX IF EXISTS ix_users_tenant_id;
DROP INDEX IF EXISTS ux_users_tenant_username;
DROP INDEX IF EXISTS ux_users_tenant_email;
DROP TABLE IF EXISTS users;
