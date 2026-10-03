package ai

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	aiv1 "github.com/ieuanign/agent-repo-template/backend/gen/ai/v1"
)

// reportTimeout bounds the wait for ai; an earlier deadline on the caller's context wins.
const reportTimeout = 30 * time.Second

// Report logs ai's service and version, or a warning when ai cannot be reached in time.
func Report(ctx context.Context, conn *grpc.ClientConn, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(ctx, reportTimeout)
	defer cancel()
	// WaitForReady rides out ai starting after backend, until the deadline.
	info, err := aiv1.NewInfoServiceClient(conn).GetInfo(ctx, &aiv1.GetInfoRequest{}, grpc.WaitForReady(true))
	if err != nil {
		log.Warn("ai unreachable at startup", zap.Error(err))
		return
	}
	log.Info("ai reachable", zap.String("service", info.GetService()), zap.String("version", info.GetVersion()))
}
