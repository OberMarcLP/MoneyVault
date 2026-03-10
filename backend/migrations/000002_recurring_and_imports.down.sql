DROP TABLE IF EXISTS import_jobs;
DROP TABLE IF EXISTS recurring_transactions;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'));
