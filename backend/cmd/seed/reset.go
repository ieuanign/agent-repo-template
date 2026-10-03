package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
)

// maintenanceDB always exists, and a database cannot be dropped over its own connection.
const maintenanceDB = "postgres"

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Close() error
}

type execOpenFunc func(database.Config) (execer, error)

func openExecer(cfg database.Config) (execer, error) {
	db, err := database.Open(cfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// reset drops and recreates the configured database, leaving it empty for migrate.
func reset(ctx context.Context, log *zap.Logger, cfg database.Config, openExec execOpenFunc) error {
	name := cfg.Name
	cfg.Name = maintenanceDB
	db, err := openExec(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	quoted := pgx.Identifier{name}.Sanitize()
	// FORCE ends backend's pooled connections, which would otherwise block the drop.
	if _, err := db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoted+" WITH (FORCE)"); err != nil {
		return fmt.Errorf("dropping database %s: %w", name, err)
	}
	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+quoted); err != nil {
		return fmt.Errorf("creating database %s: %w", name, err)
	}
	log.Info("database reset", zap.String("database", name))
	return nil
}
