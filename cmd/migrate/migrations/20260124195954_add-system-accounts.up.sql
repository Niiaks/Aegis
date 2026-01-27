-- Update wallet type constraint to allow external and platform types
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_type_check;
ALTER TABLE wallets ADD CONSTRAINT wallets_type_check CHECK (type IN ('holding', 'settlement', 'revenue', 'external', 'platform', 'seller'));

-- First, insert system user
INSERT INTO users (id, platform_id, psp_id, name, email, created_at, updated_at) VALUES 
  ('00000000-0000-0000-0000-000000000000', 'SYSTEM', 'SYSTEM', 'System Account', 'system@aegis.internal', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Then, insert system wallets (referencing the system user)
INSERT INTO wallets (id, user_id, type, balance, locked_balance, currency, created_at, updated_at) VALUES 
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'external', 0, 0, 'GHS', NOW(), NOW()),
  ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'platform', 0, 0, 'GHS', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;