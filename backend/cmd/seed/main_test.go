package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
)

type ctxKey struct{}

// fakeSeeder appends its name to calls and fails with err.
type fakeSeeder struct {
	name  string
	err   error
	calls *[]string
	ctxOK *bool
}

func (f fakeSeeder) Seed(ctx context.Context) error {
	*f.calls = append(*f.calls, f.name)
	if ctx.Value(ctxKey{}) != "run" {
		*f.ctxOK = false
	}
	return f.err
}

var pg = []string{"POSTGRES_HOST=postgres", "POSTGRES_USER=app", "POSTGRES_PASSWORD=s3cret", "POSTGRES_DB=template"}

func envFor(appEnv string) []string { return append([]string{"APP_ENV=" + appEnv}, pg...) }

type runCase struct {
	name      string
	args      []string
	environ   []string
	seeders   []string
	failing   string // the seeder that errors
	wantCode  int
	wantCalls []string
	wantOpen  bool
	wantOut   []string
}

func TestRun(t *testing.T) {
	tests := []runCase{
		{name: "staging refused", environ: envFor("staging"), seeders: []string{"a"}, wantCode: 1, wantOut: []string{"APP_ENV", "staging"}},
		{name: "production refused", environ: envFor("production"), seeders: []string{"a"}, wantCode: 1, wantOut: []string{"production"}},
		{name: "APP_ENV unset", environ: pg, seeders: []string{"a"}, wantCode: 1, wantOut: []string{"APP_ENV: required"}},
		{name: "unknown argument", args: []string{"-sideways"}, environ: envFor("dev"), seeders: []string{"a"}, wantCode: 2, wantOut: []string{"usage"}},
		{name: "extra argument after -reset", args: []string{"-reset", "x"}, environ: envFor("dev"), seeders: []string{"a"}, wantCode: 2, wantOut: []string{"usage"}},
		{name: "seeds in order", environ: envFor("dev"), seeders: []string{"a", "b", "c"}, wantOpen: true, wantCalls: []string{"a", "b", "c"}},
		{name: "stops at first error", environ: envFor("dev"), seeders: []string{"a", "b", "c"}, failing: "b", wantCode: 1,
			wantOpen: true, wantCalls: []string{"a", "b"}, wantOut: []string{"seeding b", "boom"}},
		{name: "empty registry", environ: envFor("dev"), wantOpen: true, wantOut: []string{"nothing to seed"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			ctxOK := true
			var opened *database.Config
			open := func(cfg database.Config) (*sqlx.DB, error) {
				opened = &cfg
				return database.Open(cfg)
			}
			reg := func(*sqlx.DB, *zap.Logger) []entry {
				var es []entry
				for _, n := range tt.seeders {
					var err error
					if n == tt.failing {
						err = errors.New("boom")
					}
					es = append(es, entry{n, fakeSeeder{n, err, &calls, &ctxOK}})
				}
				return es
			}
			var stdout, stderr bytes.Buffer
			ctx := context.WithValue(context.Background(), ctxKey{}, "run")
			openExec := func(database.Config) (execer, error) {
				t.Error("maintenance connection opened outside -reset")
				return &fakeExecer{}, nil
			}
			code := run(ctx, append([]string{"seed"}, tt.args...), tt.environ, &stdout, &stderr, open, openExec, reg)
			out := stdout.String() + stderr.String()

			if code != tt.wantCode {
				t.Errorf("exit %d, want %d; output:\n%s", code, tt.wantCode, out)
			}
			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("seeders called %v, want %v", calls, tt.wantCalls)
			}
			if !ctxOK {
				t.Error("a seeder did not receive the run's context")
			}
			switch {
			case tt.wantOpen && opened == nil:
				t.Error("opener not called")
			case tt.wantOpen && opened.Name != "template":
				t.Errorf("opened database %q, want template", opened.Name)
			case !tt.wantOpen && opened != nil:
				t.Error("opener called; want a refusal before any connection")
			}
			for _, s := range tt.wantOut {
				if !strings.Contains(out, s) {
					t.Errorf("output lacks %q:\n%s", s, out)
				}
			}
			if strings.Contains(out, "s3cret") {
				t.Errorf("output holds the password:\n%s", out)
			}
		})
	}
}

// fakeExecer records each statement and fails the one starting with failOn.
type fakeExecer struct {
	stmts  []string
	failOn string
	closed bool
}

func (f *fakeExecer) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	f.stmts = append(f.stmts, query)
	if f.failOn != "" && strings.HasPrefix(query, f.failOn) {
		return nil, errors.New("exec boom")
	}
	return nil, nil
}

func (f *fakeExecer) Close() error { f.closed = true; return nil }

func TestRunReset(t *testing.T) {
	quotedEnv := append([]string{"APP_ENV=dev", `POSTGRES_DB=in"s`}, pg[:3]...)
	tests := []struct {
		name      string
		environ   []string
		failOn    string
		wantCode  int
		wantOpen  bool
		wantStmts []string
		wantOut   []string
	}{
		{name: "staging refused", environ: envFor("staging"), wantCode: 1, wantOut: []string{"staging"}},
		{name: "production refused", environ: envFor("production"), wantCode: 1, wantOut: []string{"production"}},
		{name: "APP_ENV unset", environ: pg, wantCode: 1, wantOut: []string{"APP_ENV: required"}},
		{name: "drops and recreates", environ: envFor("dev"), wantOpen: true, wantStmts: []string{
			`DROP DATABASE IF EXISTS "template" WITH (FORCE)`,
			`CREATE DATABASE "template"`,
		}},
		{name: "quotes the name", environ: quotedEnv, wantOpen: true, wantStmts: []string{
			`DROP DATABASE IF EXISTS "in""s" WITH (FORCE)`,
			`CREATE DATABASE "in""s"`,
		}},
		{name: "drop fails", environ: envFor("dev"), failOn: "DROP", wantCode: 1, wantOpen: true,
			wantStmts: []string{`DROP DATABASE IF EXISTS "template" WITH (FORCE)`}, wantOut: []string{"dropping", "exec boom"}},
		{name: "create fails", environ: envFor("dev"), failOn: "CREATE", wantCode: 1, wantOpen: true,
			wantStmts: []string{`DROP DATABASE IF EXISTS "template" WITH (FORCE)`, `CREATE DATABASE "template"`},
			wantOut:   []string{"creating", "exec boom"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeExecer{failOn: tt.failOn}
			var opened *database.Config
			openExec := func(cfg database.Config) (execer, error) {
				opened = &cfg
				return fake, nil
			}
			open := func(database.Config) (*sqlx.DB, error) {
				t.Error("configured database opened by -reset")
				return nil, errors.New("unexpected")
			}
			reg := func(*sqlx.DB, *zap.Logger) []entry {
				t.Error("registry built by -reset")
				return nil
			}
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"seed", "-reset"}, tt.environ, &stdout, &stderr, open, openExec, reg)
			out := stdout.String() + stderr.String()

			if code != tt.wantCode {
				t.Errorf("exit %d, want %d; output:\n%s", code, tt.wantCode, out)
			}
			switch {
			case tt.wantOpen && opened == nil:
				t.Error("maintenance connection not opened")
			case tt.wantOpen && opened.Name != "postgres":
				t.Errorf("connected to %q, want the maintenance database postgres", opened.Name)
			case !tt.wantOpen && opened != nil:
				t.Error("connection opened; want a refusal before any connection")
			}
			if tt.wantOpen && !fake.closed {
				t.Error("maintenance connection not closed")
			}
			if !slices.Equal(fake.stmts, tt.wantStmts) {
				t.Errorf("statements %q, want %q", fake.stmts, tt.wantStmts)
			}
			for _, s := range tt.wantOut {
				if !strings.Contains(out, s) {
					t.Errorf("output lacks %q:\n%s", s, out)
				}
			}
			if strings.Contains(out, "s3cret") {
				t.Errorf("output holds the password:\n%s", out)
			}
		})
	}
}
