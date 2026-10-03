package api

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/apperr"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

// statuses is the only place a failure's HTTP status is chosen.
var statuses = map[apperr.Kind]int{
	apperr.NotFound:     http.StatusNotFound,
	apperr.Unauthorized: http.StatusUnauthorized,
	apperr.Forbidden:    http.StatusForbidden,
	apperr.Validation:   http.StatusBadRequest,
	apperr.Conflict:     http.StatusConflict,
	apperr.RateLimited:  http.StatusTooManyRequests,
	apperr.Timeout:      http.StatusGatewayTimeout,
	apperr.Internal:     http.StatusInternalServerError,
}

// mapError turns any error into the envelope's status, code and message, logging INTERNAL
// outcomes only: 4xx text can echo request values.
func mapError(ctx context.Context, err error) (status int, code, message string) {
	var kind apperr.Kind
	var refined *apperr.Error
	switch {
	case errors.As(err, &refined):
		kind, code, message = refined.Kind(), refined.Code(), refined.Message()
	case errors.As(err, &kind):
		code, message = kind.Code(), kind.Message()
	}
	status, ok := statuses[kind]
	if !ok {
		kind, status = apperr.Internal, http.StatusInternalServerError
		code, message = apperr.Internal.Code(), apperr.Internal.Message()
	}
	if kind == apperr.Internal {
		logger.FromContext(ctx).Error("request failed", zap.Error(err))
	}
	return status, code, message
}
