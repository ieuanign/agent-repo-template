-- Every change here breaks the previous release's image.
BEGIN;
ALTER TABLE projects RENAME COLUMN note TO notes;
ALTER TABLE projects DROP COLUMN legacy_ref;
ALTER TABLE projects ALTER COLUMN amount TYPE numeric(12, 2);
ALTER TABLE projects ADD COLUMN owner_id bigint NOT NULL;
ALTER TABLE projects ALTER COLUMN notes SET NOT NULL;
COMMIT;
