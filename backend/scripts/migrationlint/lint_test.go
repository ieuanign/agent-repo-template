package main

import (
	"fmt"
	"testing"
	"testing/fstest"
)

const wrapped = "BEGIN;\nSELECT 1;\nCOMMIT;\n"

// tree always has a projects module, so fixtures named for it pass the module rule.
func tree(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{"internal/modules/projects/.gitkeep": {}}
	for name, src := range files {
		fsys["migrations/"+name] = &fstest.MapFile{Data: []byte(src)}
	}
	return fsys
}

func rulesFor(vs []violation, file string) map[string]bool {
	got := map[string]bool{}
	for _, v := range vs {
		if v.File == file {
			got[v.Rule] = true
		}
	}
	return got
}

func TestNames(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		wantName bool
	}{
		{"valid", "20260101000000_projects_create_project.up.sql", false},
		{"no timestamp", "projects_create_project.up.sql", true},
		{"13 digits", "2026010100000_projects_create_project.up.sql", true},
		{"15 digits", "202601010000000_projects_create_project.up.sql", true},
		{"no such date", "20260230000000_projects_create_project.up.sql", true},
		{"uppercase desc", "20260101000000_projects_Create_project.up.sql", true},
		{"hyphenated desc", "20260101000000_projects_create-project.up.sql", true},
		{"no desc", "20260101000000_projects.up.sql", true},
		{"wrong extension", "20260101000000_projects_create_project.up.psql", true},
		{"stray non-SQL file", "notes.txt", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			down := "20260101000000_projects_create_project.down.sql"
			vs, err := lintFiles(tree(map[string]string{tt.file: wrapped, down: wrapped}), nil)
			if err != nil {
				t.Fatalf("lintFiles: %v", err)
			}
			got := rulesFor(vs, tt.file)
			if got["name"] != tt.wantName {
				t.Fatalf("name violation = %t, want %t; got %v", got["name"], tt.wantName, vs)
			}
			if tt.wantName && len(got) != 1 {
				t.Fatalf("a bad name must be reported by name only, got %v", got)
			}
		})
	}
}

func TestNamesSubdirectory(t *testing.T) {
	fsys := tree(nil)
	fsys["migrations/sub/20260101000000_projects_x.up.sql"] = &fstest.MapFile{Data: []byte(wrapped)}
	vs, err := lintFiles(fsys, nil)
	if err != nil {
		t.Fatalf("lintFiles: %v", err)
	}
	if got := rulesFor(vs, "sub"); !got["name"] || len(vs) != 1 {
		t.Fatalf("want one name violation on sub, got %v", vs)
	}
}

func TestNamesSkipped(t *testing.T) {
	vs, err := lintFiles(tree(map[string]string{"migrations.go": "package migrations", ".DS_Store": ""}), nil)
	if err != nil {
		t.Fatalf("lintFiles: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("want no violations, got %v", vs)
	}
}

func TestPairModuleOrder(t *testing.T) {
	const (
		up    = "20260102000000_projects_a.up.sql"
		down  = "20260102000000_projects_a.down.sql"
		older = "20260101000000_projects_b.up.sql"
	)
	tests := []struct {
		name  string
		files []string
		main  []string
		want  map[string]string // file -> rule
	}{
		{"complete pair", []string{up, down}, nil, map[string]string{}},
		{"up without down", []string{up}, nil, map[string]string{up: "pair"}},
		{"down without up", []string{down}, nil, map[string]string{down: "pair"}},
		{"unknown module", []string{"20260102000000_tasks_a.up.sql", "20260102000000_tasks_a.down.sql"}, nil,
			map[string]string{"20260102000000_tasks_a.up.sql": "module", "20260102000000_tasks_a.down.sql": "module"}},
		{"newer than main", []string{older, "20260101000000_projects_b.down.sql", up, down}, []string{older, "20260101000000_projects_b.down.sql"}, map[string]string{}},
		{"older than main", []string{older, "20260101000000_projects_b.down.sql", up, down}, []string{up, down},
			map[string]string{older: "order", "20260101000000_projects_b.down.sql": "order"}},
		{"equal to main", []string{up, down, "20260102000000_projects_c.up.sql", "20260102000000_projects_c.down.sql"}, []string{up, down},
			map[string]string{"20260102000000_projects_c.up.sql": "order", "20260102000000_projects_c.down.sql": "order"}},
		{"main ignores non-migrations", []string{up, down}, []string{"migrations.go"}, map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{}
			for _, f := range tt.files {
				files[f] = wrapped
			}
			vs, err := lintFiles(tree(files), tt.main)
			if err != nil {
				t.Fatalf("lintFiles: %v", err)
			}
			got := map[string]string{}
			for _, v := range vs {
				got[v.File] = v.Rule
			}
			if len(vs) != len(tt.want) || fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Fatalf("got %v, want %v", vs, tt.want)
			}
		})
	}
}

func TestModulesTreeMissing(t *testing.T) {
	fsys := fstest.MapFS{"migrations/20260101000000_projects_a.up.sql": {Data: []byte(wrapped)}, "migrations/20260101000000_projects_a.down.sql": {Data: []byte(wrapped)}}
	vs, err := lintFiles(fsys, nil)
	if err != nil {
		t.Fatalf("lintFiles: %v", err)
	}
	if len(vs) != 2 || vs[0].Rule != "module" || vs[1].Rule != "module" {
		t.Fatalf("want two module violations, got %v", vs)
	}
}

func TestMigrationsFolderMissing(t *testing.T) {
	if _, err := lintFiles(fstest.MapFS{"internal/modules/projects/.gitkeep": {}}, nil); err == nil {
		t.Fatal("want an error for a missing migrations/")
	}
}
