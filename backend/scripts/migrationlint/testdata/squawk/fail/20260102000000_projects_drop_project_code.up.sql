BEGIN;
-- squawk-ignore ban-drop-column
ALTER TABLE projects DROP COLUMN code;
COMMIT;
