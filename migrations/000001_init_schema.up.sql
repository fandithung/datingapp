CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR NOT NULL UNIQUE,
    password_hash VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    bio TEXT,
    birth_date DATE NOT NULL,
    gender VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscription_features (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_features (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    feature_id UUID NOT NULL REFERENCES subscription_features(id),
    value INTEGER NOT NULL DEFAULT 0,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    status VARCHAR NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, feature_id)
);

CREATE TABLE IF NOT EXISTS profile_responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID NOT NULL REFERENCES users(id),
    to_user_id UUID NOT NULL REFERENCES users(id),
    response_type VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (response_type IN ('like', 'pass')),
    CHECK (from_user_id != to_user_id),
    UNIQUE (from_user_id, to_user_id)
);

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_profile_responses_from_user_id ON profile_responses(from_user_id);
CREATE INDEX IF NOT EXISTS idx_profile_responses_to_user_id ON profile_responses(to_user_id);

CREATE TABLE IF NOT EXISTS daily_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    usage_date DATE NOT NULL,
    response_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, usage_date),
    CHECK (response_count >= 0)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_features_user_id ON user_features(user_id);
CREATE INDEX IF NOT EXISTS idx_user_features_feature_id ON user_features(feature_id);
CREATE INDEX IF NOT EXISTS idx_profile_responses_from_user_id ON profile_responses(from_user_id);
CREATE INDEX IF NOT EXISTS idx_profile_responses_to_user_id ON profile_responses(to_user_id);
CREATE INDEX IF NOT EXISTS idx_daily_usage_user_id_date ON daily_usage(user_id, usage_date);

-- Insert default subscription features
INSERT INTO subscription_features (id, name, description) VALUES
    (uuid_generate_v4(), 'daily_responses', 'Number of daily responses allowed');