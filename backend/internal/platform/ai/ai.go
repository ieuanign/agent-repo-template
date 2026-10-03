// Package ai owns backend's one connection to ai and the startup report; modules build their own stubs from the connection.
package ai

import (
	"fmt"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

// New returns a lazy connection: nothing dials until the first call, so boot never needs ai.
// opts are applied after the package's own.
func New(cfg config.AI, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	all := append([]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, policy()...)
	all = append(all, opts...)
	conn, err := grpc.NewClient(target, all...)
	if err != nil {
		return nil, fmt.Errorf("creating the ai connection: %w", err)
	}
	return conn, nil
}
