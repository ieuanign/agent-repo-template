package logger

import (
	"crypto/rand"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

// RequestID echoes the caller's X-Request-ID, or assigns one, and stores a
// child of base carrying it, and the ID itself, in the request context.
func RequestID(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = rand.Text()
		}
		c.Header(RequestIDHeader, id)
		l := base.With(zap.String("request_id", id))
		ctx := WithRequestID(WithContext(c.Request.Context(), l), id)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
