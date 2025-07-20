-- TRIGGERS & FUNCTIONS
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP TRIGGER IF EXISTS trigger_user_details_updated_at ON user_details;
DROP FUNCTION IF EXISTS update_timestamp_on_change();

-- INDEXES
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_user_details_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;

-- TABLES (Bağımlılık sırasına göre)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS user_details;
DROP TABLE IF EXISTS users;

-- ENUM TYPES
DROP TYPE IF EXISTS auth_provider;
DROP TYPE IF EXISTS role;
DROP TYPE IF EXISTS user_status;

-- EXTENSIONS
DROP EXTENSION IF EXISTS "uuid-ossp";
