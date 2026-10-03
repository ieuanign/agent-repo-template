package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Not t.TempDir: this repository keeps throwaway files under .scratch/, never the system temp dir.
func scratchDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("no repository root above the test")
		}
		root = parent
	}
	scratch := filepath.Join(root, ".scratch")
	if err := os.MkdirAll(scratch, 0o755); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(scratch, "migrationlint-")
	if err != nil {
		t.Fatal(err)
	}
	// .scratch/ itself stays: lint.sh may be creating its own directory in it concurrently.
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// Isolated from global config and GIT_* so a user's signing, hooks or GIT_DIR cannot change a fixture repo.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	base := []string{"-c", "user.name=migrationlint", "-c", "user.email=migrationlint@example.invalid",
		"-c", "commit.gpgsign=false", "-c", "core.hooksPath=" + os.DevNull}
	cmd := exec.Command("git", append(base, args...)...)
	cmd.Dir = dir
	cmd.Env = append(withoutGitVars(os.Environ()), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func copyDir(t *testing.T, from, to string, skip string) {
	t.Helper()
	err := filepath.WalkDir(from, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		if d.IsDir() {
			if rel == skip {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(to, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(to, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A set's main/ is committed so the order rule has a baseline; the rest stays uncommitted, like a branch's new files.
func fixtureRepo(t *testing.T, set string) string {
	t.Helper()
	dir := scratchDir(t)
	migrations := filepath.Join(dir, "migrations")
	git(t, dir, "init", "-q", "-b", "main")
	if main := filepath.Join(set, "main"); exists(main) {
		copyDir(t, main, migrations, "")
		git(t, dir, "add", "migrations")
	}
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "main")
	copyDir(t, set, migrations, "main")
	copyDir(t, filepath.Join("testdata", "modules"), filepath.Join(dir, "internal", "modules"), "")
	return dir
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Fixtures use backend's real config, so they test the rules that ship rather than a copy.
var squawkConfig = filepath.Join("..", "..", ".squawk.toml")

func TestFixtures(t *testing.T) {
	tests := []struct {
		rule      string
		wantFiles []string // every failing file in the fail set
	}{
		{"order", []string{"20260101000000_projects_create_projects.down.sql", "20260101000000_projects_create_projects.up.sql"}},
		{"pair", []string{"20260101000000_projects_create_projects.up.sql", "20260102000000_projects_create_project_notes.down.sql"}},
		{"name", []string{"projects_create_projects.down.sql", "projects_create_projects.up.sql"}},
		{"module", []string{"20260101000000_tasks_create_tasks.down.sql", "20260101000000_tasks_create_tasks.up.sql"}},
		{"wrap", []string{"20260101000000_projects_create_projects.up.sql"}},
		{"squawk", []string{"20260101000000_projects_reshape_projects.up.sql", "20260102000000_projects_drop_project_code.up.sql"}},
	}
	for _, tt := range tests {
		t.Run(tt.rule+"/pass", func(t *testing.T) {
			vs, err := check(squawkConfig, fixtureRepo(t, filepath.Join("testdata", tt.rule, "pass")))
			if err != nil {
				t.Fatalf("check: %v", err)
			}
			if len(vs) != 0 {
				t.Fatalf("want no violations, got %+v", vs)
			}
		})
		t.Run(tt.rule+"/fail", func(t *testing.T) {
			vs, err := check(squawkConfig, fixtureRepo(t, filepath.Join("testdata", tt.rule, "fail")))
			if err != nil {
				t.Fatalf("check: %v", err)
			}
			files := map[string]bool{}
			for _, v := range vs {
				if v.Rule != tt.rule {
					t.Fatalf("want only %s violations, got %+v", tt.rule, v)
				}
				files[v.File] = true
			}
			var got []string
			for f := range files {
				got = append(got, f)
			}
			sort.Strings(got)
			if strings.Join(got, " ") != strings.Join(tt.wantFiles, " ") {
				t.Fatalf("failing files %v, want %v", got, tt.wantFiles)
			}
		})
	}
}

func TestRunReportsEachViolation(t *testing.T) {
	var out bytes.Buffer
	if code := run(squawkConfig, fixtureRepo(t, filepath.Join("testdata", "pair", "fail")), &out); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	for _, want := range []string{`"guard": 7`, `"rule": "pair"`, "20260101000000_projects_create_projects.up.sql", "20260102000000_projects_create_project_notes.down.sql"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %s:\n%s", want, out.String())
		}
	}
}

func TestRunPasses(t *testing.T) {
	var out bytes.Buffer
	if code := run(squawkConfig, fixtureRepo(t, filepath.Join("testdata", "pair", "pass")), &out); code != 0 {
		t.Fatalf("exit %d, want 0:\n%s", code, out.String())
	}
}

func TestFailsClosed(t *testing.T) {
	t.Run("missing main ref", func(t *testing.T) {
		dir := scratchDir(t)
		git(t, dir, "init", "-q", "-b", "trunk")
		git(t, dir, "commit", "-q", "--allow-empty", "-m", "trunk")
		copyDir(t, filepath.Join("testdata", "pair", "pass"), filepath.Join(dir, "migrations"), "")
		if _, err := check(squawkConfig, dir); err == nil {
			t.Fatal("want an error without a main ref")
		}
	})
	t.Run("missing migrations folder", func(t *testing.T) {
		dir := scratchDir(t)
		git(t, dir, "init", "-q", "-b", "main")
		git(t, dir, "commit", "-q", "--allow-empty", "-m", "main")
		if _, err := check(squawkConfig, dir); err == nil {
			t.Fatal("want an error without migrations/")
		}
	})
	t.Run("squawk not on PATH", func(t *testing.T) {
		dir := fixtureRepo(t, filepath.Join("testdata", "pair", "pass"))
		gitPath, err := exec.LookPath("git")
		if err != nil {
			t.Fatal(err)
		}
		bin := scratchDir(t)
		if err := os.Symlink(gitPath, filepath.Join(bin, "git")); err != nil {
			t.Fatal(err)
		}
		// git stays reachable, so only squawk's absence can fail the check.
		t.Setenv("PATH", bin)
		if _, err := check(squawkConfig, dir); err == nil || !strings.Contains(err.Error(), "squawk") {
			t.Fatalf("want a squawk error with ups present, got %v", err)
		}
	})
	t.Run("missing squawk config", func(t *testing.T) {
		dir := fixtureRepo(t, filepath.Join("testdata", "pair", "pass"))
		if _, err := check(filepath.Join(dir, ".squawk.toml"), dir); err == nil {
			t.Fatal("want an error without .squawk.toml")
		}
	})
}

func TestSquawkNamesEveryRule(t *testing.T) {
	vs, err := check(squawkConfig, fixtureRepo(t, filepath.Join("testdata", "squawk", "fail")))
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	lines := map[string]int{}
	for _, v := range vs {
		lines[strings.SplitN(v.Msg, ":", 2)[0]] = v.Line
	}
	want := map[string]int{"renaming-column": 3, "ban-drop-column": 4, "changing-column-type": 5, "adding-required-field": 6, "adding-not-nullable-field": 7}
	for rule, line := range want {
		if lines[rule] != line {
			t.Errorf("%s on line %d, want %d; got %+v", rule, lines[rule], line, vs)
		}
	}
}

func TestSquawkSyntaxErrorFails(t *testing.T) {
	dir := fixtureRepo(t, filepath.Join("testdata", "pair", "pass"))
	up := filepath.Join(dir, "migrations", "20260101000000_projects_create_projects.up.sql")
	if err := os.WriteFile(up, []byte("BEGIN;\nSELEC oops;\nCOMMIT;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	vs, err := check(squawkConfig, dir)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(vs) == 0 || vs[0].Rule != "squawk" || !strings.HasPrefix(vs[0].Msg, "syntax-error") {
		t.Fatalf("want a squawk syntax-error violation, got %+v", vs)
	}
}
