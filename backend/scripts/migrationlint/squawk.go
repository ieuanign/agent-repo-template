package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

const markedForm = "-- squawk-ignore <rule>[, <rule>…] -- second release: <why the old shape can go>"

// squawk ignores everything after the second "--", so the reason costs nothing at lint time.
var markedRe = regexp.MustCompile(`^\s*-- squawk-ignore [a-z0-9-]+(?:, [a-z0-9-]+)* -- second release: \S`)

// checkIgnores lets squawk be silenced only by a line marking the statement as the second, removing release.
func checkIgnores(file, src string) []violation {
	var vs []violation
	for i, l := range strings.Split(src, "\n") {
		if strings.Contains(l, "squawk-ignore") && !markedRe.MatchString(l) {
			vs = append(vs, violation{8, "squawk", file, i + 1, "an ignore must be a line of its own directly above the flagged line: " + markedForm})
		}
	}
	return vs
}

// Downs are skipped: dropping what the up made is their job, and squawk's ban-drop rules would flag it.
func runSquawk(cfg, migrations string) ([]violation, error) {
	entries, err := os.ReadDir(migrations)
	if err != nil {
		return nil, fmt.Errorf("reading migrations/: %w", err)
	}
	var ups []string
	for _, e := range entries {
		if m, ok := parseName(e.Name()); ok && m.up && !e.IsDir() {
			ups = append(ups, m.file)
		}
	}
	if len(ups) == 0 {
		return nil, nil
	}
	cmd := exec.Command("squawk", append([]string{"--config", cfg, "--reporter", "json"}, ups...)...)
	cmd.Dir = migrations
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	code := 0
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return nil, fmt.Errorf("running squawk: %w", err)
		}
		code = exit.ExitCode()
	}
	return squawkViolations(stdout.Bytes(), code, stderr.String())
}

// A failed run with no findings is an error, never a pass: squawk also exits non-zero when it cannot run.
func squawkViolations(out []byte, code int, stderr string) ([]violation, error) {
	var findings []struct {
		File    string `json:"file"`
		Line    int    `json:"line"`
		Rule    string `json:"rule_name"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(out, &findings); err != nil {
		return nil, fmt.Errorf("parsing squawk's report (exit %d): %w: %s", code, err, strings.TrimSpace(stderr))
	}
	if code != 0 && len(findings) == 0 {
		return nil, fmt.Errorf("squawk exited %d without findings: %s", code, strings.TrimSpace(stderr))
	}
	var vs []violation
	for _, f := range findings {
		msg := fmt.Sprintf("%s: %s; ship it over two releases, and mark the removing one with %s", f.Rule, f.Message, markedForm)
		if f.Rule == "syntax-error" {
			msg = fmt.Sprintf("%s: squawk cannot parse it: %s", f.Rule, f.Message)
		}
		// squawk's JSON lines are 0-based.
		vs = append(vs, violation{8, "squawk", path.Base(f.File), f.Line + 1, msg})
	}
	return vs, nil
}
