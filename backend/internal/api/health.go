package api

import (
	"context"
	"time"

	"github.com/ieuanign/agent-repo-template/backend/gen/openapi"
)

// GetHealth is liveness: it checks no dependency.
func (s *Server) GetHealth(context.Context, openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	return openapi.GetHealth200JSONResponse{Version: s.version, Environment: openapi.HealthEnvironment(s.env)}, nil
}

// readyTimeout stays under the healthcheck client's timeout.
const readyTimeout = 2 * time.Second

// GetHealthReady answers 503, with code null, while the database does not respond: not-ready is
// this operation's declared answer, not an error. The ping error is dropped so nothing names the database.
func (s *Server) GetHealthReady(ctx context.Context, _ openapi.GetHealthReadyRequestObject) (openapi.GetHealthReadyResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, readyTimeout)
	defer cancel()
	if s.db.PingContext(ctx) != nil {
		return openapi.GetHealthReady503Response{}, nil
	}
	return openapi.GetHealthReady200Response{}, nil
}
