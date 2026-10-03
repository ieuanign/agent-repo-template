package main

import "testing"

func TestWrap(t *testing.T) {
	tests := []struct {
		name string
		src  string
		fail bool
	}{
		{"empty file", "", true},
		{"comments only", "-- nothing\n/* here */\n", true},
		{"missing BEGIN", "SELECT 1;\nCOMMIT;\n", true},
		{"missing COMMIT", "BEGIN;\nSELECT 1;\n", true},
		{"statement before BEGIN", "SELECT 1;\nBEGIN;\nSELECT 2;\nCOMMIT;\n", true},
		{"statement after COMMIT", "BEGIN;\nSELECT 1;\nCOMMIT;\nSELECT 2;\n", true},
		{"nested BEGIN", "BEGIN;\nBEGIN;\nSELECT 1;\nCOMMIT;\n", true},
		{"mid-file COMMIT", "BEGIN;\nSELECT 1;\nCOMMIT;\nSELECT 2;\nCOMMIT;\n", true},
		{"mid-file END", "BEGIN;\nSELECT 1;\nEND;\nCOMMIT;\n", true},
		{"mid-file ROLLBACK", "BEGIN;\nSELECT 1;\nROLLBACK;\nCOMMIT;\n", true},
		{"mid-file START TRANSACTION", "BEGIN;\nstart  transaction;\nCOMMIT;\n", true},
		{"mid-file ABORT", "BEGIN;\nABORT;\nCOMMIT;\n", true},
		{"mid-file PREPARE TRANSACTION", "BEGIN;\nPREPARE TRANSACTION 'x';\nCOMMIT;\n", true},
		{"BEGIN WORK is not exactly BEGIN", "BEGIN WORK;\nSELECT 1;\nCOMMIT;\n", true},
		{"unterminated block comment hides COMMIT", "BEGIN;\nSELECT 1;\n/* /* */ COMMIT;\n", true},
		{"wrapped", "BEGIN;\nSELECT 1;\nCOMMIT;\n", false},
		{"lowercase", "begin;\nselect 1;\ncommit;\n", false},
		{"COMMIT without trailing semicolon", "BEGIN;\nSELECT 1;\nCOMMIT", false},
		{"comments outside", "-- up\nBEGIN;\nSELECT 1;\nCOMMIT;\n-- done\n", false},
		{"ROLLBACK TO SAVEPOINT", "BEGIN;\nSAVEPOINT s;\nSELECT 1;\nROLLBACK TO SAVEPOINT s;\nCOMMIT;\n", false},
		{"END inside dollar quotes", "BEGIN;\nDO $$ BEGIN PERFORM 1; END; $$;\nCOMMIT;\n", false},
		{"END inside tagged dollar quotes", "BEGIN;\nCREATE FUNCTION f() RETURNS int AS $fn$ BEGIN RETURN 1; END; $fn$ LANGUAGE plpgsql;\nCOMMIT;\n", false},
		{"semicolons in strings", "BEGIN;\nSELECT 'a; COMMIT;', E'it\\'s; END;', \"x;y\";\nCOMMIT;\n", false},
		{"semicolons in comments", "BEGIN;\n-- COMMIT;\nSELECT 1 /* ; END; /* nested; */ ROLLBACK; */;\nCOMMIT;\n", false},
		{"doubled quote in string", "BEGIN;\nSELECT 'it''s; COMMIT;';\nCOMMIT;\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vs := checkWrap("f.up.sql", tt.src)
			if (len(vs) > 0) != tt.fail {
				t.Fatalf("violations %v, want fail=%t", vs, tt.fail)
			}
			for _, v := range vs {
				if v.Rule != "wrap" || v.Guard != 7 {
					t.Fatalf("want guard 7 wrap, got %+v", v)
				}
			}
		})
	}
}

func TestWrapLine(t *testing.T) {
	vs := checkWrap("f.up.sql", "BEGIN;\nSELECT 1;\n\n  ROLLBACK;\nCOMMIT;\n")
	if len(vs) != 1 || vs[0].Line != 4 {
		t.Fatalf("want one violation on line 4, got %+v", vs)
	}
}

func TestWrapAppliesToUpAndDown(t *testing.T) {
	up, down := "20260101000000_projects_a.up.sql", "20260101000000_projects_a.down.sql"
	vs, err := lintFiles(tree(map[string]string{up: "", down: "SELECT 1;"}), nil)
	if err != nil {
		t.Fatalf("lintFiles: %v", err)
	}
	if !rulesFor(vs, up)["wrap"] || !rulesFor(vs, down)["wrap"] {
		t.Fatalf("want wrap on both files, got %v", vs)
	}
}
