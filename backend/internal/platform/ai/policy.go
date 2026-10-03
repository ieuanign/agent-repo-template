package ai

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

const defaultTimeout = 10 * time.Second

// Only UNAVAILABLE retries: DEADLINE_EXCEEDED means ai is already slow, RESOURCE_EXHAUSTED that it sheds load.
const serviceConfig = `{"methodConfig": [{
	"name": [{}],
	"retryPolicy": {
		"maxAttempts": 3,
		"initialBackoff": "0.1s",
		"backoffMultiplier": 2,
		"maxBackoff": "1s",
		"retryableStatusCodes": ["UNAVAILABLE"]
	}
}]}`

func policy() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithDefaultServiceConfig(serviceConfig),
		// Otherwise a _grpc_config TXT record on ai's host would replace the retry policy.
		grpc.WithDisableServiceConfig(),
		grpc.WithChainUnaryInterceptor(defaultDeadline, requestID),
	}
}

// defaultDeadline sets one only when the caller has none; a method-config timeout would also cap longer ones.
func defaultDeadline(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func requestID(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if id, ok := logger.RequestIDFromContext(ctx); ok && printableASCII(id) {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", id)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// printableASCII guards the call: gRPC fails it with INTERNAL on such metadata, and HTTP lets tabs and 0x80+ through.
func printableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7e {
			return false
		}
	}
	return true
}
