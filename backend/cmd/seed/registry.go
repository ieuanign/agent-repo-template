package main

import (
	"context"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// seeder is what each module's Seed exposes. Seeders write through their module's service with
// fixed natural keys and fake data, skipping what exists, because every `make dev` re-runs them.
type seeder interface {
	Seed(ctx context.Context) error
}

// entry names a module by its directory under internal/modules.
type entry struct {
	name string
	s    seeder
}

type registryFunc func(db *sqlx.DB, log *zap.Logger) []entry

// registry lists modules in cmd/server's construction order, so a seeder may use earlier modules' data.
func registry(db *sqlx.DB, log *zap.Logger) []entry {
	return nil
}

// checkRegistered fails naming every module directory directly under fsys that has no registered seeder.
func checkRegistered(fsys fs.FS, names []string) error {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading the modules tree: %w", err)
	}
	var missing []string
	for _, e := range entries {
		if e.IsDir() && !slices.Contains(names, e.Name()) {
			missing = append(missing, e.Name())
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		return fmt.Errorf("modules with no registered seeder: %s", strings.Join(missing, ", "))
	}
	return nil
}
