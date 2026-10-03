-- Projects, touched by a trigger on every update.
BEGIN;

/* outer /* nested */ still a comment; COMMIT; */
CREATE TABLE projects (
    id bigint PRIMARY KEY,
    note text NOT NULL DEFAULT 'BEGIN; COMMIT;',
    updated_at timestamptz
);

CREATE FUNCTION projects_touch() RETURNS trigger AS $fn$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$fn$ LANGUAGE plpgsql;

CREATE TRIGGER projects_touch BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION projects_touch();

DO $$
BEGIN
    RAISE NOTICE 'projects ready';
END
$$;

COMMIT;
-- End of migration.
