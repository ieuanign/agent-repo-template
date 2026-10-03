package main

import "testing"

func TestSquawkOutput(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		code    int
		want    int
		wantErr bool
	}{
		{"clean", "[]", 0, 0, false},
		{"findings", `[{"file":"a.up.sql","line":0,"rule_name":"ban-drop-column","message":"m"}]`, 1, 1, false},
		{"failed without findings", "[]", 1, 0, true},
		{"failed with no output", "", 1, 0, true},
		{"unparseable", "oops", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vs, err := squawkViolations([]byte(tt.out), tt.code, "")
			if (err != nil) != tt.wantErr || len(vs) != tt.want {
				t.Fatalf("got %+v, %v", vs, err)
			}
			if tt.want > 0 && (vs[0].Line != 1 || vs[0].Guard != 8 || vs[0].Rule != "squawk" || vs[0].File != "a.up.sql") {
				t.Fatalf("want guard 8 squawk on a.up.sql line 1, got %+v", vs[0])
			}
		})
	}
}
