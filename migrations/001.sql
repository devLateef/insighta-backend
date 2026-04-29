-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id            VARCHAR(36) PRIMARY KEY,
    github_id     VARCHAR(64) UNIQUE NOT NULL,
    username      VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    avatar_url    VARCHAR(512),
    role          VARCHAR(32) NOT NULL DEFAULT 'analyst',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Refresh tokens table (server-side invalidation)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          SERIAL PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(512) UNIQUE NOT NULL,
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Profiles table (Stage 2 spec — name is UNIQUE)
CREATE TABLE IF NOT EXISTS profiles (
    id                   VARCHAR(36) PRIMARY KEY,
    name                 VARCHAR(255) NOT NULL UNIQUE,
    gender               VARCHAR(32),
    gender_probability   FLOAT,
    age                  INTEGER,
    age_group            VARCHAR(32),
    country_id           VARCHAR(2),
    country_name         VARCHAR(255),
    country_probability  FLOAT,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for filtering performance
CREATE INDEX IF NOT EXISTS idx_profiles_gender             ON profiles(gender);
CREATE INDEX IF NOT EXISTS idx_profiles_country_id         ON profiles(country_id);
CREATE INDEX IF NOT EXISTS idx_profiles_age                ON profiles(age);
CREATE INDEX IF NOT EXISTS idx_profiles_age_group          ON profiles(age_group);
CREATE INDEX IF NOT EXISTS idx_profiles_gender_prob        ON profiles(gender_probability);
CREATE INDEX IF NOT EXISTS idx_profiles_country_prob       ON profiles(country_probability);
CREATE INDEX IF NOT EXISTS idx_profiles_created_at         ON profiles(created_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id      ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash   ON refresh_tokens(token_hash);
