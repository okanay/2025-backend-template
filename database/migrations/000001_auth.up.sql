-- EXTENSIONS
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- HELPER ENUM TYPES
CREATE TYPE user_status AS ENUM ('Active', 'Suspended', 'Deleted');
CREATE TYPE role AS ENUM ('User', 'Editor', 'Admin');
CREATE TYPE auth_provider AS ENUM ('credentials', 'google', 'facebook', 'twitter', 'apple', 'microsoft', 'github', 'linkedin');

-- USER TABLE: Kimlik ve rolü tutar.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    auth_provider auth_provider DEFAULT 'credentials' NOT NULL,
    hashed_password TEXT,
    role role DEFAULT 'User' NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    status user_status DEFAULT 'Active' NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    last_login TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- USER DETAILS TABLE: Opsiyonel profil verilerini barındırır.
CREATE TABLE IF NOT EXISTS user_details (
    id TEXT PRIMARY KEY,
    user_id TEXT UNIQUE NOT NULL,
    provider_id TEXT,
    display_name TEXT,
    first_name TEXT,
    last_name TEXT,
    avatar_url TEXT,
    phone_e164 TEXT UNIQUE,
    phone_country_code TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- PERMISSIONS TABLE: Sistemdeki tüm potansiyel izinlerin sözlüğü. Yeni yetkiler migration ile buraya eklenir.
CREATE TABLE IF NOT EXISTS permissions (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL, -- Örn: "post:delete:any"
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- USER_PERMISSIONS TABLE: Hangi kullanıcının hangi izne sahip olduğunu bağlayan dinamik tablo.
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- REFRESH TOKEN TABLE: Güvenli oturum yönetimini sağlar.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    user_email TEXT,
    token TEXT UNIQUE NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days',
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    last_used_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    revoked_reason TEXT,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- INDEXES: Hızlı sorgular için kritik öneme sahiptir.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_user_details_user_id ON user_details(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);

-- TRIGGERS & FUNCTIONS: 'updated_at' gibi alanları otomatik günceller.
CREATE OR REPLACE FUNCTION update_timestamp_on_change()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamp_on_change();
CREATE TRIGGER trigger_user_details_updated_at BEFORE UPDATE ON user_details FOR EACH ROW EXECUTE FUNCTION update_timestamp_on_change();
