package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"slices"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// NewMigrator runs fsys's migrations on a connection of its own, which its Close closes.
func NewMigrator(cfg Config, fsys fs.FS) (*migrate.Migrate, error) {
	src, err := iofs.New(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("reading the embedded migrations: %w", err)
	}
	db, err := sql.Open("pgx", cfg.connString())
	if err != nil {
		return nil, fmt.Errorf("opening the migration connection: %w", err)
	}
	drv, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connecting the migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "pgx5", drv)
	if err != nil {
		_ = drv.Close()
		return nil, fmt.Errorf("creating the migrator: %w", err)
	}
	return m, nil
}

// LatestVersion returns the newest embedded version; ok is false when none is embedded.
func LatestVersion(fsys fs.FS) (v uint, ok bool, err error) {
	vs, err := versions(fsys)
	if err != nil || len(vs) == 0 {
		return 0, false, err
	}
	return vs[len(vs)-1], true, nil
}

// PrevVersion returns the embedded version before v, or -1 when v is the first;
// ok is false when v is not embedded.
func PrevVersion(fsys fs.FS, v uint) (prev int, ok bool, err error) {
	vs, err := versions(fsys)
	if err != nil {
		return 0, false, err
	}
	i, found := slices.BinarySearch(vs, v)
	switch {
	case !found:
		return 0, false, nil
	case i == 0:
		return -1, true, nil
	}
	return int(vs[i-1]), true, nil
}

// versions skips names that are not migrations, as iofs does.
func versions(fsys fs.FS) ([]uint, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("listing the embedded migrations: %w", err)
	}
	var vs []uint
	for _, e := range entries {
		m, err := source.Parse(e.Name())
		if e.IsDir() || err != nil {
			continue
		}
		vs = append(vs, m.Version)
	}
	slices.Sort(vs)
	return slices.Compact(vs), nil
}
