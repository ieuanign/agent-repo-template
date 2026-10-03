BEGIN;
ALTER TABLE projects ALTER COLUMN notes DROP NOT NULL;
ALTER TABLE projects DROP COLUMN owner_id;
ALTER TABLE projects ALTER COLUMN amount TYPE bigint;
ALTER TABLE projects ADD COLUMN legacy_ref text;
ALTER TABLE projects RENAME COLUMN notes TO note;
COMMIT;
