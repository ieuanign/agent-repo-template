package database

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestConnStringRoundTrips(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{"plain", Config{Host: "postgres", Port: 5432, User: "app", Password: "secret", Name: "template"}},
		{"special characters", Config{Host: "db.internal", Port: 6543, User: "a@b:c", Password: `p@ss:/?#%w rd`, Name: "my/db?#"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pgx.ParseConfig(tt.cfg.connString())
			if err != nil {
				t.Fatalf("ParseConfig: %v", err)
			}
			if got.Host != tt.cfg.Host || Port(got.Port) != tt.cfg.Port || got.User != tt.cfg.User ||
				got.Password != tt.cfg.Password || got.Database != tt.cfg.Name {
				t.Fatalf("round trip lost a value: host=%q port=%d user=%q db=%q (password match %t)",
					got.Host, got.Port, got.User, got.Database, got.Password == tt.cfg.Password)
			}
		})
	}
}

func TestOpenDoesNotDial(t *testing.T) {
	// 10.255.255.1 is unroutable, so a dial would hang rather than fail fast.
	cfg := Config{Host: "10.255.255.1", Port: 5432, User: "app", Password: "secret", Name: "template"}
	start := time.Now()
	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if d := time.Since(start); d > time.Second {
		t.Fatalf("Open took %s; it must not dial", d)
	}
	if n := db.Stats().OpenConnections; n != 0 {
		t.Fatalf("Open made %d connections; want 0", n)
	}
	if got := db.DriverName(); got != "pgx" {
		t.Fatalf("driver %q; want pgx for $n placeholders", got)
	}
}
