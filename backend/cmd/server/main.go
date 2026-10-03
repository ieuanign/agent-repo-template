// Command server serves backend's HTTP API.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/gen/openapi"
	"github.com/ieuanign/agent-repo-template/backend/internal/api"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/ai"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/buildinfo"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
	"github.com/ieuanign/agent-repo-template/backend/migrations"
)

// shutdownTimeout stays under Compose's 30 s stop grace period.
const shutdownTimeout = 20 * time.Second

// versioner is the part of *migrate.Migrate the boot check uses.
type versioner interface {
	Version() (version uint, dirty bool, err error)
	Close() (source error, database error)
}

type openFunc func(database.Config, fs.FS) (versioner, error)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	open := func(cfg database.Config, fsys fs.FS) (versioner, error) { return database.NewMigrator(cfg, fsys) }
	err := run(ctx, os.Args, os.Environ(), os.Stdout, migrations.FS, open)
	stop()
	if err != nil {
		msg := "server exited"
		if isHealthcheck(os.Args) {
			msg = "healthcheck failed"
		}
		// APP_ENV may be what failed, so the encoder cannot depend on it.
		logger.New(config.Production, os.Stdout).Error(msg, zap.Error(err))
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, environ []string, out io.Writer, fsys fs.FS, open openFunc) error {
	if isHealthcheck(args) {
		return healthcheck(ctx, environ)
	}
	cfg, err := config.Parse(environ)
	if err != nil {
		return err
	}
	log := logger.New(cfg.AppEnv, out)
	defer func() { _ = log.Sync() }()

	if err := checkSchema(log, cfg.Database, fsys, open); err != nil {
		return err
	}

	spec, err := openapi.GetSpec()
	if err != nil {
		return err
	}
	specJSON, err := openapi.GetSpecJSON()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		return err
	}
	// Deferred, so it runs after the drain below: in-flight requests still use the pool.
	defer func() { _ = db.Close() }()
	// Deferred after the pool, so it closes after the drain and before the pool.
	aiConn, err := ai.New(cfg.AI)
	if err != nil {
		return err
	}
	defer func() { _ = aiConn.Close() }()

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		return err
	}
	log.Info("starting", zap.String("service", "backend"),
		zap.String("version", buildinfo.Version), zap.String("commit", buildinfo.Commit),
		zap.String("addr", ln.Addr().String()))

	rctx, stopReport := context.WithCancel(ctx)
	reported := make(chan struct{})
	go func() {
		defer close(reported)
		ai.Report(rctx, aiConn, log)
	}()
	// out is not safe for concurrent writes, so no line of run's own follows until the report returns.
	waitReport := func() { stopReport(); <-reported }
	defer waitReport()

	srv := &http.Server{
		Handler:           api.NewRouter(api.NewServer(buildinfo.Version, cfg.AppEnv, db), spec, log, api.Docs(cfg.AppEnv, specJSON)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}
	waitReport()
	log.Info("shutting down")
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil {
		return err
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("stopped")
	return nil
}

// checkSchema refuses a schema it cannot verify; this is the one database failure that stops boot.
func checkSchema(log *zap.Logger, cfg database.Config, fsys fs.FS, open openFunc) error {
	m, err := open(cfg, fsys)
	if err != nil {
		return fmt.Errorf("checking the schema version: %w", err)
	}
	v, dirty, err := m.Version()
	if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
		log.Warn("closing the migrator", zap.NamedError("source", srcErr), zap.NamedError("database", dbErr))
	}
	applied := !errors.Is(err, migrate.ErrNilVersion)
	if err != nil && applied {
		return fmt.Errorf("reading the schema version: %w", err)
	}
	if dirty {
		return fmt.Errorf("database is dirty at version %d: run migrate up for the recovery command", v)
	}
	latest, embedded, err := database.LatestVersion(fsys)
	if err != nil {
		return err
	}
	switch {
	case embedded && (!applied || v < latest):
		return fmt.Errorf("database schema is behind: at %d, embedded %d; run migrate up", v, latest)
	case applied && (!embedded || v > latest):
		log.Warn("database schema is ahead of this build", zap.Uint("database", v), zap.Uint("embedded", latest))
	}
	return nil
}
