package database

import (
	"testing"
	"testing/fstest"
)

// embedded builds an embedded folder holding an up/down pair per name, plus a Go file iofs must skip.
func embedded(names ...string) fstest.MapFS {
	fsys := fstest.MapFS{"migrations.go": {Data: []byte("package migrations")}}
	for _, n := range names {
		fsys[n+".up.sql"] = &fstest.MapFile{}
		fsys[n+".down.sql"] = &fstest.MapFile{}
	}
	return fsys
}

var three = embedded("20260101000000_a_one", "20260201000000_a_two", "20260301000000_b_three")

func TestLatestVersion(t *testing.T) {
	tests := []struct {
		name   string
		fsys   fstest.MapFS
		want   uint
		wantOK bool
	}{
		{"empty", embedded(), 0, false},
		{"one", embedded("20260101000000_a_one"), 20260101000000, true},
		{"three", three, 20260301000000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := LatestVersion(tt.fsys)
			if err != nil {
				t.Fatalf("LatestVersion: %v", err)
			}
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("got (%d, %t), want (%d, %t)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestPrevVersion(t *testing.T) {
	tests := []struct {
		name   string
		v      uint
		want   int
		wantOK bool
	}{
		{"first", 20260101000000, -1, true},
		{"middle", 20260201000000, 20260101000000, true},
		{"last", 20260301000000, 20260201000000, true},
		{"not embedded", 20260401000000, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := PrevVersion(three, tt.v)
			if err != nil {
				t.Fatalf("PrevVersion: %v", err)
			}
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("got (%d, %t), want (%d, %t)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
