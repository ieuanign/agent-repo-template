package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/golang-migrate/migrate/v4"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
)

func TestRunConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		names   []string
	}{
		{"missing APP_ENV", nil, []string{"APP_ENV"}},
		{"empty APP_ENV", []string{"APP_ENV="}, []string{"APP_ENV"}},
		{"invalid APP_ENV", []string{"APP_ENV=prod"}, []string{"APP_ENV"}},
		{"invalid PORT", []string{"APP_ENV=dev", "PORT=x"}, []string{"PORT"}},
		{"both bad", []string{"PORT=x"}, []string{"APP_ENV", "PORT"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(context.Background(), nil, tt.environ, &bytes.Buffer{}, none, mustNotOpen(t))
			if err == nil {
				t.Fatal("run: want error, got nil")
			}
			for _, n := range tt.names {
				if !strings.Contains(err.Error(), n) {
					t.Errorf("error %q does not name %s", err, n)
				}
			}
		})
	}
}

func TestRunCancelledContextShutsDownCleanly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	// An unroutable PostgreSQL host: boot must not need the database.
	environ := []string{"APP_ENV=dev", "PORT=0", "POSTGRES_HOST=10.255.255.1",
		"POSTGRES_USER=app", "POSTGRES_PASSWORD=pw", "POSTGRES_DB=template", "AI_HOST=10.255.255.1"}
	if err := run(ctx, nil, environ, &out, none, (&fakeVersioner{}).open); err != nil {
		t.Fatalf("run: %v", err)
	}
	var startup []string
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Contains(line, `"service": "backend"`) {
			startup = append(startup, line)
		}
	}
	if len(startup) != 1 {
		t.Fatalf("want one startup line, got %d in:\n%s", len(startup), out.String())
	}
	for _, key := range []string{`"version":`, `"commit":`} {
		if !strings.Contains(startup[0], key) {
			t.Errorf("startup line %q lacks %s", startup[0], key)
		}
	}
	// The report's warning in the buffer proves it ran and returned before run did.
	if !strings.Contains(out.String(), "ai unreachable at startup") {
		t.Errorf("no ai report warning in:\n%s", out.String())
	}
}

// fakeVersioner stands in for the migrator; applied false means no migration has run.
type fakeVersioner struct {
	applied    bool
	version    uint
	dirty      bool
	versionErr error
	closed     bool
}

func (f *fakeVersioner) Version() (uint, bool, error) {
	switch {
	case f.versionErr != nil:
		return 0, false, f.versionErr
	case !f.applied:
		return 0, false, migrate.ErrNilVersion
	}
	return f.version, f.dirty, nil
}

func (f *fakeVersioner) Close() (error, error) {
	f.closed = true
	return nil, nil
}

func (f *fakeVersioner) open(database.Config, fs.FS) (versioner, error) { return f, nil }

func mustNotOpen(t *testing.T) openFunc {
	return func(database.Config, fs.FS) (versioner, error) {
		t.Error("migrator factory called")
		return nil, errors.New("not expected")
	}
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
	none = embedded()
	two  = embedded("20260101000000_a_one", "20260201000000_b_two")
)

func TestRunSchemaCheck(t *testing.T) {
	tests := []struct {
		name     string
		fsys     fs.FS
		v        *fakeVersioner
		openErr  error
		wantErr  string
		wantWarn bool
	}{
		{name: "both empty", fsys: none, v: &fakeVersioner{}},
		{name: "equal", fsys: two, v: &fakeVersioner{applied: true, version: 20260201000000}},
		{name: "ahead", fsys: two, v: &fakeVersioner{applied: true, version: 20260301000000}, wantWarn: true},
		{name: "ahead of an empty source", fsys: none, v: &fakeVersioner{applied: true, version: 20260101000000}, wantWarn: true},
		{name: "behind", fsys: two, v: &fakeVersioner{applied: true, version: 20260101000000}, wantErr: "behind"},
		{name: "none applied", fsys: two, v: &fakeVersioner{}, wantErr: "behind"},
		{name: "dirty", fsys: two, v: &fakeVersioner{applied: true, version: 20260201000000, dirty: true}, wantErr: "dirty"},
		{name: "unreadable", fsys: two, v: &fakeVersioner{versionErr: errors.New("refused")}, wantErr: "refused"},
		{name: "cannot open", fsys: two, openErr: errors.New("no route"), wantErr: "no route"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			open := func(database.Config, fs.FS) (versioner, error) {
				if tt.openErr != nil {
					return nil, tt.openErr
				}
				return tt.v, nil
			}
			environ := []string{"APP_ENV=dev", "PORT=0", "POSTGRES_HOST=10.255.255.1",
				"POSTGRES_USER=app", "POSTGRES_PASSWORD=s3cret", "POSTGRES_DB=template", "AI_HOST=10.255.255.1"}
			var out bytes.Buffer
			err := run(ctx, nil, environ, &out, tt.fsys, open)
			started := strings.Contains(out.String(), `"service": "backend"`)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("run: err = %v, want one naming %q", err, tt.wantErr)
				}
				if started {
					t.Errorf("listener opened despite the refusal:\n%s", out.String())
				}
			} else {
				if err != nil {
					t.Fatalf("run: %v", err)
				}
				if !started {
					t.Errorf("server did not start:\n%s", out.String())
				}
			}
			warned := strings.Contains(out.String(), "database schema is ahead of this build")
			if warned != tt.wantWarn {
				t.Errorf("warned = %v, want %v:\n%s", warned, tt.wantWarn, out.String())
			}
			if tt.wantWarn {
				for _, s := range []string{strconv.FormatUint(uint64(tt.v.version), 10), `"embedded"`} {
					if !strings.Contains(out.String(), s) {
						t.Errorf("warning lacks %s:\n%s", s, out.String())
					}
				}
			}
			if tt.v != nil && !tt.v.closed {
				t.Error("migrator not closed")
			}
			if strings.Contains(out.String(), "s3cret") {
				t.Errorf("output holds the password:\n%s", out.String())
			}
		})
	}
}
