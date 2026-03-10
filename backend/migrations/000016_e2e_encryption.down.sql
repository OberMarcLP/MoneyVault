ALTER TABLE users
  DROP COLUMN IF EXISTS e2e_enabled,
  DROP COLUMN IF EXISTS e2e_encrypted_dek,
  DROP COLUMN IF EXISTS e2e_kek_salt;
