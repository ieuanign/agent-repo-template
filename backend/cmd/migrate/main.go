// Command migrate applies backend's embedded SQL migrations.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
	"github.com/ieuanign/agent-repo-template/backend/migrations"
)

const usage = "usage: migrate up | version | down | redo | force -version N -confirm <POSTGRES_DB>"

// migrator is the part of *migrate.Migrate this command uses.
type migrator interface {
	Version() (version uint, dirty bool, err error)
	Up() error
	Steps(n int) error
	Force(version int) error
	Close() (source error, database error)
}

type openFunc func(database.Config, fs.FS) (migrator, error)

type settings struct {
	AppEnv   config.AppEnv `env:"APP_ENV,required,notEmpty"`
	Database database.Config
}

func main() {
	open := func(cfg database.Config, fsys fs.FS) (migrator, error) { return database.NewMigrator(cfg, fsys) }
	os.Exit(run(os.Args, os.Environ(), os.Stdout, os.Stderr, migrations.FS, open))
}

func run(args []string, environ []string, stdout, stderr io.Writer, fsys fs.FS, open openFunc) int {
	// APP_ENV may be what failed, so the encoder cannot depend on it.
	early := logger.New(config.Production, stderr)
	if len(args) < 2 {
		early.Error(usage)
		return 2
	}
	cmd := args[1]
	switch cmd {
	case "up", "version", "down", "redo", "force":
	default:
		early.Error(usage)
		return 2
	}
	cfg, err := config.ParseAs[settings](environ)
	if err != nil {
		early.Error("migrate refused", zap.Error(err))
		return 1
	}
	log := logger.New(cfg.AppEnv, stdout)
	defer func() { _ = log.Sync() }()

	// An allowlist, so staging, production and any future environment fail closed.
	if (cmd == "down" || cmd == "redo") && cfg.AppEnv != config.Dev {
		log.Error("migrate "+cmd+" refused: it runs only when APP_ENV is dev", zap.String("app_env", string(cfg.AppEnv)))
		return 1
	}

	var forceTo int
	if cmd == "force" {
		if forceTo, err = forceFlags(args[2:], cfg.Database.Name); err != nil {
			log.Error("migrate force refused", zap.Error(err))
			return 1
		}
	}

	m, err := open(cfg.Database, fsys)
	if err != nil {
		log.Error("migrate failed", zap.Error(err))
		return 1
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			log.Warn("closing the migrator", zap.NamedError("source", srcErr), zap.NamedError("database", dbErr))
		}
	}()

	switch cmd {
	case "up":
		err = up(log, m, fsys, cfg.Database.Name)
	case "version":
		err = version(log, m)
	case "down", "redo":
		err = stepDown(log, m, cmd == "redo")
	case "force":
		err = force(log, m, forceTo)
	}
	if err != nil {
		log.Error("migrate "+cmd+" failed", zap.Error(err))
		return 1
	}
	return 0
}

// readVersion reports applied false when no migration has run.
func readVersion(m migrator) (v uint, applied, dirty bool, err error) {
	v, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, fmt.Errorf("reading the schema version: %w", err)
	}
	return v, true, dirty, nil
}

func up(log *zap.Logger, m migrator, fsys fs.FS, dbName string) error {
	v, applied, dirty, err := readVersion(m)
	if err != nil {
		return err
	}
	if dirty {
		return dirtyErr(log, fsys, v, dbName)
	}
	latest, ok, err := database.LatestVersion(fsys)
	if err != nil {
		return err
	}
	// Up fails on an empty source and on a database ahead of it, so neither reaches it.
	if !ok || (applied && v >= latest) {
		log.Info("nothing to do", zap.Uint("database", v), zap.Bool("applied", applied), zap.Uint("embedded", latest))
		return nil
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrating up: %w", err)
	}
	log.Info("migrated up", zap.Uint("version", latest))
	return nil
}

// dirtyErr names the force command; each migration commits alone, so its predecessor is fully applied.
func dirtyErr(log *zap.Logger, fsys fs.FS, v uint, dbName string) error {
	prev, ok, err := database.PrevVersion(fsys, v)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("database is dirty at version %d, which this image has not embedded", v)
	}
	log.Error("database is dirty: fix what the failed migration left, then force the version before it",
		zap.Uint("dirty", v), zap.String("fix", fmt.Sprintf("make migrate-force version=%d confirm=%s", prev, dbName)))
	return fmt.Errorf("database is dirty at version %d", v)
}

func version(log *zap.Logger, m migrator) error {
	v, applied, dirty, err := readVersion(m)
	if err != nil {
		return err
	}
	if !applied {
		log.Info("no migration applied")
		return nil
	}
	log.Info("schema version", zap.Uint("version", v), zap.Bool("dirty", dirty))
	return nil
}

// stepDown reverts exactly one migration; Down would revert them all.
func stepDown(log *zap.Logger, m migrator, redo bool) error {
	v, applied, _, err := readVersion(m)
	if err != nil {
		return err
	}
	if !applied {
		log.Info("nothing to revert: no migration applied")
		return nil
	}
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("reverting version %d: %w", v, err)
	}
	log.Info("reverted", zap.Uint("version", v))
	if !redo {
		return nil
	}
	if err := m.Steps(1); err != nil {
		return fmt.Errorf("reapplying version %d: %w", v, err)
	}
	log.Info("reapplied", zap.Uint("version", v))
	return nil
}

// forceFlags demands the database's own name, so a force cannot land on the wrong database by habit.
func forceFlags(args []string, dbName string) (int, error) {
	fl := flag.NewFlagSet("force", flag.ContinueOnError)
	fl.SetOutput(io.Discard)
	v := fl.String("version", "", "")
	confirm := fl.String("confirm", "", "")
	if err := fl.Parse(args); err != nil {
		return 0, fmt.Errorf("%w; %s", err, usage)
	}
	if *v == "" {
		return 0, errors.New("-version is required: an integer, -1 or more")
	}
	n, err := strconv.Atoi(*v)
	if err != nil || n < -1 {
		return 0, fmt.Errorf("-version %q must be an integer, -1 or more", *v)
	}
	if *confirm != dbName {
		return 0, errors.New("-confirm must be exactly the configured POSTGRES_DB")
	}
	return n, nil
}

func force(log *zap.Logger, m migrator, to int) error {
	if err := version(log, m); err != nil {
		return err
	}
	if err := m.Force(to); err != nil {
		return fmt.Errorf("forcing version %d: %w", to, err)
	}
	log.Info("forced", zap.Int("version", to))
	return nil
}
