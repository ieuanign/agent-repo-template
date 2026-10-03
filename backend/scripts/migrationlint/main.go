// Command migrationlint checks backend's migration files against the forward-only migration rules.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
	"github.com/ieuanign/agent-repo-template/backend/internal/platform/logger"
)

func main() {
	os.Exit(run(".squawk.toml", ".", os.Stderr))
}

// run takes its paths and writer, and returns the exit code, so tests drive it against fixture repositories.
func run(cfg, dir string, out io.Writer) int {
	log := logger.New(config.Dev, out)
	defer func() { _ = log.Sync() }()
	vs, err := check(cfg, dir)
	if err != nil {
		log.Error("migration lint could not run", zap.Error(err))
		return 1
	}
	for _, v := range vs {
		fields := []zap.Field{zap.Int("guard", v.Guard), zap.String("rule", v.Rule), zap.String("file", v.File)}
		if v.Line > 0 {
			fields = append(fields, zap.Int("line", v.Line))
		}
		log.Error(v.Msg, fields...)
	}
	if len(vs) > 0 {
		log.Error("migration lint failed", zap.Int("violations", len(vs)))
		return 1
	}
	return 0
}

func check(cfg, dir string) ([]violation, error) {
	cfg, err := filepath.Abs(cfg)
	if err != nil {
		return nil, fmt.Errorf("resolving the squawk config: %w", err)
	}
	if _, err := os.Stat(cfg); err != nil {
		return nil, fmt.Errorf("reading the squawk config: %w", err)
	}
	mainNames, err := mainMigrations(dir)
	if err != nil {
		return nil, err
	}
	vs, err := lintFiles(os.DirFS(dir), mainNames)
	if err != nil {
		return nil, err
	}
	sv, err := runSquawk(cfg, filepath.Join(dir, "migrations"))
	if err != nil {
		return nil, err
	}
	return append(vs, sv...), nil
}

// A missing main ref is an error, never an empty baseline: an empty one would pass every order check.
func mainMigrations(dir string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-z", "--name-only", "main", "--", "migrations/")
	cmd.Dir = dir
	cmd.Env = withoutGitVars(os.Environ())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("reading main's migrations: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var names []string
	for _, p := range strings.Split(string(b), "\x00") {
		if p != "" {
			names = append(names, path.Base(p))
		}
	}
	return names, nil
}

// withoutGitVars drops GIT_* so a hook's GIT_DIR cannot point git away from the repository holding dir.
func withoutGitVars(env []string) []string {
	var out []string
	for _, kv := range env {
		if !strings.HasPrefix(kv, "GIT_") {
			out = append(out, kv)
		}
	}
	return out
}
