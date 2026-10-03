package main

import "testing"

func TestIgnores(t *testing.T) {
	tests := []struct {
		name string
		line string
		fail bool
	}{
		{"bare", "-- squawk-ignore ban-drop-column", true},
		{"file-wide", "-- squawk-ignore-file", true},
		{"file-wide marked", "-- squawk-ignore-file ban-drop-column -- second release: gone", true},
		{"block comment", "/* squawk-ignore ban-drop-column -- second release: gone */", true},
		{"trailing a statement", "ALTER TABLE projects DROP COLUMN legacy; -- squawk-ignore ban-drop-column -- second release: gone", true},
		{"empty reason", "-- squawk-ignore ban-drop-column -- second release: ", true},
		{"no rule", "-- squawk-ignore -- second release: gone", true},
		{"one rule", "-- squawk-ignore ban-drop-column -- second release: nothing has read legacy since v3", false},
		{"several rules", "-- squawk-ignore ban-drop-column, renaming-column -- second release: v3 reads notes", false},
		{"indented", "    -- squawk-ignore ban-drop-column -- second release: gone", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "BEGIN;\n" + tt.line + "\nALTER TABLE projects DROP COLUMN legacy;\nCOMMIT;\n"
			vs := checkIgnores("f.up.sql", src)
			if (len(vs) > 0) != tt.fail {
				t.Fatalf("violations %v, want fail=%t", vs, tt.fail)
			}
			for _, v := range vs {
				if v.Guard != 8 || v.Rule != "squawk" || v.Line != 2 {
					t.Fatalf("want guard 8 squawk on line 2, got %+v", v)
				}
			}
		})
	}
}

func TestIgnoresAppliesToUpAndDown(t *testing.T) {
	up, down := "20260101000000_projects_a.up.sql", "20260101000000_projects_a.down.sql"
	bare := "BEGIN;\n-- squawk-ignore ban-drop-column\nSELECT 1;\nCOMMIT;\n"
	vs, err := lintFiles(tree(map[string]string{up: bare, down: bare}), nil)
	if err != nil {
		t.Fatalf("lintFiles: %v", err)
	}
	if !rulesFor(vs, up)["squawk"] || !rulesFor(vs, down)["squawk"] {
		t.Fatalf("want squawk on both files, got %v", vs)
	}
}
