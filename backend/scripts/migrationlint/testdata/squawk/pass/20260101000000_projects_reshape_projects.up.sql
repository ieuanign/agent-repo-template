-- The removing release: the previous image no longer reads any old shape here.
BEGIN;
-- squawk-ignore renaming-column -- second release: the previous image reads notes
ALTER TABLE projects RENAME COLUMN note TO notes;
-- squawk-ignore ban-drop-column -- second release: nothing has read legacy_ref since it was replaced
ALTER TABLE projects DROP COLUMN legacy_ref;
-- squawk-ignore changing-column-type -- second release: the previous image writes numeric amounts
ALTER TABLE projects ALTER COLUMN amount TYPE numeric(12, 2);
-- squawk-ignore adding-required-field -- second release: the previous image sets owner_id
ALTER TABLE projects ADD COLUMN owner_id bigint NOT NULL;
-- squawk-ignore adding-not-nullable-field -- second release: the previous image always writes notes
ALTER TABLE projects ALTER COLUMN notes SET NOT NULL;
CREATE INDEX projects_owner_id_idx ON projects (owner_id);
COMMIT;
