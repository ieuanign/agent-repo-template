package main

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"time"
)

type violation struct {
	Guard int
	Rule  string
	File  string
	Line  int // 0 when unknown
	Msg   string
}

type migration struct {
	file      string
	timestamp string
	module    string
	stem      string
	up        bool
}

var nameRe = regexp.MustCompile(`^(\d{14})_([a-z][a-z0-9]*)_([a-z0-9]+(?:_[a-z0-9]+)*)\.(up|down)\.sql$`)

const nameForm = "<YYYYMMDDhhmmss UTC>_<module>_<snake_case desc>.{up,down}.sql"

// mainNames are base names, not paths, so they compare directly with the entries of migrations/.
func lintFiles(backend fs.FS, mainNames []string) ([]violation, error) {
	entries, err := fs.ReadDir(backend, "migrations")
	if err != nil {
		return nil, fmt.Errorf("reading migrations/: %w", err)
	}
	var vs []violation
	var ms []migration
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, ".") || (!e.IsDir() && strings.HasSuffix(n, ".go")) {
			continue
		}
		if e.IsDir() {
			vs = append(vs, violation{7, "name", n, 0, "subdirectory: golang-migrate reads only the top level of migrations/"})
			continue
		}
		m, ok := parseName(n)
		if !ok {
			vs = append(vs, violation{7, "name", n, 0, "golang-migrate skips it silently unless it is named " + nameForm})
			continue
		}
		ms = append(ms, m)
	}
	modules, err := readModules(backend)
	if err != nil {
		return nil, err
	}
	vs = append(vs, checkPairs(ms)...)
	vs = append(vs, checkOrder(ms, mainNames)...)
	for _, m := range ms {
		if !modules[m.module] {
			vs = append(vs, violation{7, "module", m.file, 0, fmt.Sprintf("module %q is not a directory under internal/modules/", m.module)})
		}
		src, err := fs.ReadFile(backend, "migrations/"+m.file)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", m.file, err)
		}
		vs = append(vs, checkWrap(m.file, string(src))...)
		vs = append(vs, checkIgnores(m.file, string(src))...)
	}
	return vs, nil
}

func readModules(backend fs.FS) (map[string]bool, error) {
	entries, err := fs.ReadDir(backend, "internal/modules")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading internal/modules/: %w", err)
	}
	modules := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			modules[e.Name()] = true
		}
	}
	return modules, nil
}

func checkPairs(ms []migration) []violation {
	have := map[string]bool{}
	for _, m := range ms {
		have[m.file] = true
	}
	var vs []violation
	for _, m := range ms {
		partner := m.stem + ".up.sql"
		if m.up {
			partner = m.stem + ".down.sql"
		}
		if !have[partner] {
			vs = append(vs, violation{7, "pair", m.file, 0, "missing its partner " + partner})
		}
	}
	return vs
}

// checkOrder compares whole names against main: golang-migrate silently skips a new version below the applied one.
func checkOrder(ms []migration, mainNames []string) []violation {
	onMain := map[string]bool{}
	latest := ""
	for _, n := range mainNames {
		onMain[n] = true
		if m, ok := parseName(n); ok && m.timestamp > latest {
			latest = m.timestamp
		}
	}
	var vs []violation
	for _, m := range ms {
		if !onMain[m.file] && m.timestamp <= latest {
			vs = append(vs, violation{6, "order", m.file, 0, "timestamp must be newer than main's latest migration " + latest})
		}
	}
	return vs
}

func parseName(n string) (migration, bool) {
	g := nameRe.FindStringSubmatch(n)
	if g == nil {
		return migration{}, false
	}
	if _, err := time.Parse("20060102150405", g[1]); err != nil {
		return migration{}, false
	}
	return migration{file: n, timestamp: g[1], module: g[2], stem: strings.TrimSuffix(strings.TrimSuffix(n, ".sql"), "."+g[4]), up: g[4] == "up"}, true
}
