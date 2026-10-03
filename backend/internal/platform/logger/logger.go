// Package logger builds backend's zap logger and carries it through request contexts.
package logger

import (
	"context"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

// New logs to out: human-readable console in dev, JSON everywhere else.
func New(env config.AppEnv, out io.Writer) *zap.Logger {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	level := zapcore.InfoLevel
	if env == config.Dev {
		enc = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		level = zapcore.DebugLevel
	}
	return zap.New(zapcore.NewCore(enc, zapcore.AddSync(out), level))
}

type ctxKey struct{}

// WithContext returns ctx carrying l.
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns the request logger, or a no-op logger when none was stored.
// A *gin.Context reaches the request context only with ContextWithFallback on.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	}
	return zap.NewNop()
}

type requestIDKey struct{}

// WithRequestID returns ctx carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the request ID and whether one was stored.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}
