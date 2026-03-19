BEGIN;

ALTER TYPE asset_status RENAME VALUE 'for_repair' TO 'waiting_for_repair';

ALTER TABLE users RENAME COLUMN phone_no TO phone_number;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

DROP INDEX IF EXISTS idx_unique_email;

CREATE UNIQUE INDEX idx_users_email_unique
    ON users (LOWER(email))
    WHERE archived_at IS NULL;

ALTER TABLE assets RENAME COLUMN serial_no TO serial_number;
ALTER TABLE assets RENAME COLUMN assigned_on TO assigned_at;
ALTER TABLE assets RENAME COLUMN returned_on TO returned_at;
ALTER TABLE assets RENAME COLUMN assigned_by_id TO assigned_by;
ALTER TABLE assets RENAME COLUMN archived_by TO archived_by_user;
ALTER TABLE assets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

ALTER TABLE user_session RENAME TO user_sessions;

ALTER TABLE laptop RENAME TO laptops;
ALTER TABLE keyboard RENAME TO keyboards;
ALTER TABLE mouse RENAME TO mice;

ALTER TYPE user_role RENAME VALUE 'project-manager' TO 'project_manager';
ALTER TYPE user_role RENAME VALUE 'asset-manager' TO 'asset_manager';
ALTER TYPE user_type RENAME VALUE 'full-time' TO 'full_time';

COMMIT;