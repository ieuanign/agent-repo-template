package api

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/gen/openapi"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/apperr"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

// NewRouter builds the engine cmd/server serves and every test drives.
func NewRouter(impl openapi.StrictServerInterface, spec *openapi3.T, base *zap.Logger, opts ...Option) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.ContextWithFallback = true
	r.Use(logger.RequestID(base), recovery())

	h := openapi.NewStrictHandlerWithOptions(impl, nil, openapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  func(c *gin.Context, _ error) { fail(c, apperr.Validation) },
		HandlerErrorFunc:         fail,
		ResponseErrorHandlerFunc: fail,
	})
	// Validated on this group alone: unknown paths still reach NoRoute, and later non-spec routes skip it.
	openapi.RegisterHandlersWithOptions(r.Group("/api", envelope(), validator(spec)), h, openapi.GinServerOptions{
		ErrorHandler: func(c *gin.Context, _ error, _ int) { fail(c, apperr.Validation) },
	})
	r.NoRoute(envelope(), func(c *gin.Context) { fail(c, apperr.NotFound) })
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// fail records err for the envelope, which alone chooses the status and writes the body.
func fail(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}

// recovery replaces gin's, which writes the panic outside zap.
func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if p := recover(); p != nil {
				logger.FromContext(c).Error("panic", zap.Any("panic", p), zap.Stack("stack"))
				// Logged above, so mapError's logging is skipped; a streamed response cannot be replaced.
				if !c.Writer.Written() {
					writeFailure(c, statuses[apperr.Internal], apperr.Internal.Code(), apperr.Internal.Message())
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
