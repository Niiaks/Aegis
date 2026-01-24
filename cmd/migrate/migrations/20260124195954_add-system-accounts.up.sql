-- First, insert system user
INSERT INTO users (id, platform_id, psp_id, name, email, created_at, updated_at) VALUES 
  ('00000000-0000-0000-0000-000000000000', 'SYSTEM', 'SYSTEM', 'System Account', 'system@aegis.internal', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Then, insert system wallets (referencing the system user)
INSERT INTO wallets (id, user_id, type, balance, locked_balance, created_at, updated_at) VALUES 
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'external', 0, 0, NOW(), NOW()),
  ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'platform', 0, 0, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;