// Package api composes backend's HTTP surface on the generated strict server.
package api

import (
	"context"

	"github.com/ieuanign/agent-repo-template/backend/gen/openapi"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

var _ openapi.StrictServerInterface = (*Server)(nil)

// Pinger is what readiness checks; *sqlx.DB satisfies it.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Server implements every generated operation.
type Server struct {
	version string
	env     config.AppEnv
	db      Pinger
}

func NewServer(version string, env config.AppEnv, db Pinger) *Server {
	return &Server{version: version, env: env, db: db}
}
