package main

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"go.uber.org/zap"
)

func dirs(names ...string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for _, n := range names {
		fsys[n] = &fstest.MapFile{Mode: fs.ModeDir}
	}
	return fsys
}

func TestCheckRegistered(t *testing.T) {
	tests := []struct {
		name    string
		fsys    fs.FS
		names   []string
		wantErr []string // nil: must pass
		notErr  []string
	}{
		{name: "all registered", fsys: dirs("accounts", "tasks"), names: []string{"accounts", "tasks"}},
		{name: "one unregistered", fsys: dirs("accounts", "tasks"), names: []string{"accounts"}, wantErr: []string{"tasks"}, notErr: []string{"accounts"}},
		{name: "two unregistered sorted", fsys: dirs("tasks", "projects", "accounts"), names: []string{"accounts"}, wantErr: []string{"projects, tasks"}},
		{name: "root file ignored", fsys: fstest.MapFS{".gitkeep": {}, "accounts": {Mode: fs.ModeDir}}, names: []string{"accounts"}},
		{name: "nested dirs are not modules", fsys: fstest.MapFS{"accounts/internal/seed/seed.go": {}}, names: []string{"accounts"}},
		{name: "empty tree", fsys: fstest.MapFS{}, names: nil},
		{name: "missing root", fsys: mustSub(t, fstest.MapFS{}, "modules"), wantErr: []string{"modules tree"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkRegistered(tt.fsys, tt.names)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("checkRegistered: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("checkRegistered: want error, got nil")
			}
			for _, s := range tt.wantErr {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("error %q lacks %q", err, s)
				}
			}
			for _, s := range tt.notErr {
				if strings.Contains(err.Error(), s) {
					t.Errorf("error %q names registered %q", err, s)
				}
			}
		})
	}
}

func mustSub(t *testing.T, fsys fs.FS, dir string) fs.FS {
	t.Helper()
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func TestRealModulesAreRegistered(t *testing.T) {
	var names []string
	for _, e := range registry(nil, zap.NewNop()) {
		names = append(names, e.name)
	}
	if err := checkRegistered(os.DirFS("../../internal/modules"), names); err != nil {
		t.Fatal(err)
	}
}
