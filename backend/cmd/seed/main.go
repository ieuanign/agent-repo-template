// Command seed loads every module's dev data; it runs only when APP_ENV is dev.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

const usage = "usage: seed [-reset]"

type openFunc func(database.Config) (*sqlx.DB, error)

type settings struct {
	AppEnv   config.AppEnv `env:"APP_ENV,required,notEmpty"`
	Database database.Config
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	code := run(ctx, os.Args, os.Environ(), os.Stdout, os.Stderr, database.Open, openExecer, registry)
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, environ []string, stdout, stderr io.Writer, open openFunc, openExec execOpenFunc, build registryFunc) int {
	// APP_ENV may be what failed, so the encoder cannot depend on it.
	early := logger.New(config.Production, stderr)
	resetMode := len(args) == 2 && args[1] == "-reset"
	if len(args) > 1 && !resetMode {
		early.Error(usage)
		return 2
	}
	cfg, err := config.ParseAs[settings](environ)
	if err != nil {
		early.Error("seed refused", zap.Error(err))
		return 1
	}
	log := logger.New(cfg.AppEnv, stdout)
	defer func() { _ = log.Sync() }()

	// An allowlist, so staging, production and any future environment fail closed.
	if cfg.AppEnv != config.Dev {
		log.Error("seed refused: it runs only when APP_ENV is dev", zap.String("app_env", string(cfg.AppEnv)))
		return 1
	}

	if resetMode {
		if err := reset(ctx, log, cfg.Database, openExec); err != nil {
			log.Error("reset failed", zap.Error(err))
			return 1
		}
		return 0
	}
	if err := seed(ctx, log, cfg.Database, open, build); err != nil {
		log.Error("seed failed", zap.Error(err))
		return 1
	}
	return 0
}

func seed(ctx context.Context, log *zap.Logger, cfg database.Config, open openFunc, build registryFunc) error {
	db, err := open(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	entries := build(db, log)
	if len(entries) == 0 {
		log.Info("nothing to seed: no module is registered")
		return nil
	}
	for _, e := range entries {
		if err := e.s.Seed(ctx); err != nil {
			return fmt.Errorf("seeding %s: %w", e.name, err)
		}
		log.Info("seeded", zap.String("module", e.name))
	}
	return nil
}
