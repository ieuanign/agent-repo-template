BEGIN;

DROP TRIGGER projects_touch ON projects;
DROP FUNCTION projects_touch();
DROP TABLE projects;

COMMIT;
