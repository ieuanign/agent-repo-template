package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

// healthcheckTimeout sits between readiness's ping deadline and the container check's timeout.
const healthcheckTimeout = 3 * time.Second

func isHealthcheck(args []string) bool { return len(args) > 1 && args[1] == "healthcheck" }

// healthcheck probes readiness on loopback; distroless has no curl for the container check.
func healthcheck(ctx context.Context, environ []string) error {
	cfg, err := config.Parse(environ)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, healthcheckTimeout)
	defer cancel()
	url := "http://127.0.0.1:" + strconv.Itoa(cfg.Port) + "/api/health/ready"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building the readiness request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling readiness: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("readiness answered %d", resp.StatusCode)
	}
	return nil
}
