CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_id VARCHAR(255) NOT NULL,
    psp_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_platform_psp_unique UNIQUE (platform_id, psp_id)
);

CREATE INDEX idx_users_platform_id ON users(platform_id);
CREATE INDEX idx_users_psp_id ON users(psp_id);
CREATE INDEX idx_users_email ON users(email);
