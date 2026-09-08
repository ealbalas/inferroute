ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_hash_fast TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key_hash_fast ON api_keys(key_hash_fast);
UPDATE api_keys SET key_hash_fast = NULL WHERE key_hash_fast IS NULL;
