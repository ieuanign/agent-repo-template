BEGIN;

CREATE TABLE projects (
    id bigint PRIMARY KEY
);

COMMIT;

CREATE INDEX projects_id_idx ON projects (id);
