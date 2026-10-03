package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/golang-migrate/migrate/v4"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
)

// fakeMigrator records its calls; applied false means no migration has run.
type fakeMigrator struct {
	applied    bool
	version    uint
	dirty      bool
	versionErr error
	upErr      error
	calls      []string
}

func (f *fakeMigrator) Version() (uint, bool, error) {
	f.calls = append(f.calls, "version")
	switch {
	case f.versionErr != nil:
		return 0, false, f.versionErr
	case !f.applied:
		return 0, false, migrate.ErrNilVersion
	}
	return f.version, f.dirty, nil
}

func (f *fakeMigrator) Up() error {
	f.calls = append(f.calls, "up")
	return f.upErr
}

func (f *fakeMigrator) Steps(n int) error {
	f.calls = append(f.calls, fmt.Sprintf("steps %d", n))
	return nil
}

func (f *fakeMigrator) Force(v int) error {
	f.calls = append(f.calls, fmt.Sprintf("force %d", v))
	return nil
}

func (f *fakeMigrator) Close() (error, error) {
	f.calls = append(f.calls, "close")
	return nil, nil
}

func embedded(names ...string) fstest.MapFS {
	fsys := fstest.MapFS{"migrations.go": {Data: []byte("package migrations")}}
	for _, n := range names {
		fsys[n+".up.sql"] = &fstest.MapFile{}
		fsys[n+".down.sql"] = &fstest.MapFile{}
	}
	return fsys
}

var (
	none  = embedded()
	three = embedded("20260101000000_a_one", "20260201000000_a_two", "20260301000000_b_three")
	pg    = []string{"POSTGRES_HOST=postgres", "POSTGRES_USER=app", "POSTGRES_PASSWORD=s3cret", "POSTGRES_DB=template"}
)

func envFor(appEnv string) []string { return append([]string{"APP_ENV=" + appEnv}, pg...) }

type result struct {
	code   int
	out    string
	opened bool
}

func runWith(args []string, environ []string, fsys fs.FS, m *fakeMigrator) result {
	var stdout, stderr bytes.Buffer
	var opened bool
	open := func(cfg database.Config, got fs.FS) (migrator, error) {
		opened = true
		return m, nil
	}
	code := run(append([]string{"migrate"}, args...), environ, &stdout, &stderr, fsys, open)
	return result{code, stdout.String() + stderr.String(), opened}
}

type runCase struct {
	name      string
	args      []string
	environ   []string
	fsys      fs.FS
	m         *fakeMigrator
	wantCode  int
	wantCalls []string // nil: the factory must not be called
	wantOut   []string
	notOut    []string
}

func check(t *testing.T, tests []runCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			if m == nil {
				m = &fakeMigrator{}
			}
			fsys := tt.fsys
			if fsys == nil {
				fsys = three
			}
			r := runWith(tt.args, tt.environ, fsys, m)
			if (r.code == 0) != (tt.wantCode == 0) {
				t.Errorf("exit %d, want %d; output:\n%s", r.code, tt.wantCode, r.out)
			}
			if tt.wantCalls == nil {
				if r.opened {
					t.Errorf("factory called; want a refusal before any connection")
				}
			} else if !slices.Equal(m.calls, tt.wantCalls) {
				t.Errorf("calls %v, want %v", m.calls, tt.wantCalls)
			}
			for _, s := range tt.wantOut {
				if !strings.Contains(r.out, s) {
					t.Errorf("output lacks %q:\n%s", s, r.out)
				}
			}
			for _, s := range append(tt.notOut, "s3cret") {
				if strings.Contains(r.out, s) {
					t.Errorf("output holds %q:\n%s", s, r.out)
				}
			}
		})
	}
}

func TestRunConfigAndUsage(t *testing.T) {
	check(t, []runCase{
		{name: "no subcommand", environ: envFor("dev"), wantCode: 2, wantOut: []string{"usage"}},
		{name: "unknown subcommand", args: []string{"sideways"}, environ: envFor("dev"), wantCode: 2, wantOut: []string{"usage"}},
		{name: "config missing", args: []string{"up"}, wantCode: 1,
			wantOut: []string{"APP_ENV", "POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}},
		{name: "invalid APP_ENV", args: []string{"version"}, environ: envFor("prod"), wantCode: 1, wantOut: []string{"APP_ENV"}},
	})
}

func TestRunUp(t *testing.T) {
	check(t, []runCase{
		{name: "both empty", args: []string{"up"}, environ: envFor("production"), fsys: none,
			wantCalls: []string{"version", "close"}, wantOut: []string{"nothing to do"}},
		{name: "equal", args: []string{"up"}, environ: envFor("production"),
			m: &fakeMigrator{applied: true, version: 20260301000000}, wantCalls: []string{"version", "close"}, wantOut: []string{"nothing to do"}},
		{name: "ahead", args: []string{"up"}, environ: envFor("production"),
			m: &fakeMigrator{applied: true, version: 20260401000000}, wantCalls: []string{"version", "close"}, wantOut: []string{"nothing to do"}},
		{name: "ahead of an empty source", args: []string{"up"}, environ: envFor("production"), fsys: none,
			m: &fakeMigrator{applied: true, version: 20260101000000}, wantCalls: []string{"version", "close"}, wantOut: []string{"nothing to do"}},
		{name: "behind", args: []string{"up"}, environ: envFor("production"),
			m: &fakeMigrator{applied: true, version: 20260101000000}, wantCalls: []string{"version", "up", "close"}},
		{name: "none applied", args: []string{"up"}, environ: envFor("production"),
			wantCalls: []string{"version", "up", "close"}},
		{name: "no change", args: []string{"up"}, environ: envFor("production"),
			m: &fakeMigrator{upErr: migrate.ErrNoChange}, wantCalls: []string{"version", "up", "close"}},
		{name: "up fails", args: []string{"up"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{upErr: errors.New("boom")}, wantCalls: []string{"version", "up", "close"}, wantOut: []string{"boom"}},
		{name: "version unreadable", args: []string{"up"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{versionErr: errors.New("refused")}, wantCalls: []string{"version", "close"}, wantOut: []string{"refused"}},
		{name: "dirty", args: []string{"up"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{applied: true, version: 20260301000000, dirty: true}, wantCalls: []string{"version", "close"},
			wantOut: []string{"make migrate-force version=20260201000000 confirm=template"}},
		{name: "first dirty", args: []string{"up"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{applied: true, version: 20260101000000, dirty: true}, wantCalls: []string{"version", "close"},
			wantOut: []string{"make migrate-force version=-1 confirm=template"}},
		{name: "dirty version not embedded", args: []string{"up"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{applied: true, version: 20260401000000, dirty: true}, wantCalls: []string{"version", "close"},
			wantOut: []string{"not embedded"}, notOut: []string{"make migrate-force"}},
	})
}

func TestRunVersion(t *testing.T) {
	check(t, []runCase{
		{name: "none applied", args: []string{"version"}, environ: envFor("production"),
			wantCalls: []string{"version", "close"}, wantOut: []string{"no migration applied"}},
		{name: "applied", args: []string{"version"}, environ: envFor("production"),
			m: &fakeMigrator{applied: true, version: 20260201000000, dirty: true}, wantCalls: []string{"version", "close"},
			wantOut: []string{"20260201000000", `"dirty":true`}},
		{name: "unreadable", args: []string{"version"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{versionErr: errors.New("refused")}, wantCalls: []string{"version", "close"}},
	})
}

func TestRunDownAndRedoGuard(t *testing.T) {
	applied := func() *fakeMigrator { return &fakeMigrator{applied: true, version: 20260201000000} }
	check(t, []runCase{
		{name: "down staging", args: []string{"down"}, environ: envFor("staging"), wantCode: 1, wantOut: []string{"dev"}},
		{name: "down production", args: []string{"down"}, environ: envFor("production"), wantCode: 1, wantOut: []string{"dev"}},
		{name: "redo staging", args: []string{"redo"}, environ: envFor("staging"), wantCode: 1, wantOut: []string{"dev"}},
		{name: "redo production", args: []string{"redo"}, environ: envFor("production"), wantCode: 1, wantOut: []string{"dev"}},
		{name: "down dev", args: []string{"down"}, environ: envFor("dev"), m: applied(),
			wantCalls: []string{"version", "steps -1", "close"}},
		{name: "redo dev", args: []string{"redo"}, environ: envFor("dev"), m: applied(),
			wantCalls: []string{"version", "steps -1", "steps 1", "close"}},
		{name: "down dev none applied", args: []string{"down"}, environ: envFor("dev"),
			wantCalls: []string{"version", "close"}, wantOut: []string{"nothing to revert"}},
		{name: "redo dev none applied", args: []string{"redo"}, environ: envFor("dev"),
			wantCalls: []string{"version", "close"}, wantOut: []string{"nothing to revert"}},
	})
}

func TestRunForceGuard(t *testing.T) {
	dirty := func() *fakeMigrator { return &fakeMigrator{applied: true, version: 20260201000000, dirty: true} }
	check(t, []runCase{
		{name: "no flags", args: []string{"force"}, environ: envFor("production"), wantCode: 1, wantOut: []string{"-version"}},
		{name: "empty flags", args: []string{"force", "-version=", "-confirm="}, environ: envFor("production"), wantCode: 1},
		{name: "missing version", args: []string{"force", "-confirm=template"}, environ: envFor("production"), wantCode: 1},
		{name: "non-integer version", args: []string{"force", "-version=abc", "-confirm=template"}, environ: envFor("production"), wantCode: 1},
		{name: "version below -1", args: []string{"force", "-version=-2", "-confirm=template"}, environ: envFor("production"), wantCode: 1},
		{name: "missing confirm", args: []string{"force", "-version=3"}, environ: envFor("production"), wantCode: 1, wantOut: []string{"-confirm"}},
		{name: "wrong confirm", args: []string{"force", "-version=3", "-confirm=wrong"}, environ: envFor("production"), wantCode: 1},
		{name: "unknown flag", args: []string{"force", "-version=3", "-confirm=template", "-yes"}, environ: envFor("production"), wantCode: 1},
		{name: "matching confirm", args: []string{"force", "-version=20260101000000", "-confirm=template"}, environ: envFor("production"), m: dirty(),
			wantCalls: []string{"version", "force 20260101000000", "close"}, wantOut: []string{"20260201000000", `"dirty":true`}},
		{name: "to no migration", args: []string{"force", "-version=-1", "-confirm=template"}, environ: envFor("dev"),
			wantCalls: []string{"version", "force -1", "close"}, wantOut: []string{"no migration applied"}},
		{name: "version unreadable", args: []string{"force", "-version=3", "-confirm=template"}, environ: envFor("production"), wantCode: 1,
			m: &fakeMigrator{versionErr: errors.New("refused")}, wantCalls: []string{"version", "close"}},
	})
}
