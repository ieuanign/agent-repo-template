package config

import (
	"strings"
	"testing"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
)

var pg = []string{"POSTGRES_HOST=postgres", "POSTGRES_USER=app", "POSTGRES_PASSWORD=pw", "POSTGRES_DB=template"}

var pgWant = database.Config{Host: "postgres", Port: 5432, User: "app", Password: "pw", Name: "template"}

var aiWant = AI{Host: "ai", Port: 50051}

// with returns pg and AI_HOST with extra appended; later entries override earlier ones.
func with(extra ...string) []string {
	return append(append(append([]string{}, pg...), "AI_HOST=ai"), extra...)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		environ  []string
		want     Config
		errNames []string
	}{
		{"defaults ports", with("APP_ENV=dev"), Config{AppEnv: Dev, Port: 8080, Database: pgWant, AI: aiWant}, nil},
		{"all set", with("APP_ENV=production", "PORT=9000", "POSTGRES_PORT=6543", "AI_PORT=6000"),
			Config{AppEnv: Production, Port: 9000, Database: database.Config{Host: "postgres", Port: 6543, User: "app", Password: "pw", Name: "template"},
				AI: AI{Host: "ai", Port: 6000}}, nil},
		{"staging", with("APP_ENV=staging"), Config{AppEnv: Staging, Port: 8080, Database: pgWant, AI: aiWant}, nil},
		{"missing APP_ENV", with(), Config{}, []string{"APP_ENV"}},
		{"empty APP_ENV", with("APP_ENV="), Config{}, []string{"APP_ENV"}},
		{"invalid APP_ENV", with("APP_ENV=prod"), Config{}, []string{"APP_ENV"}},
		{"invalid PORT", with("APP_ENV=dev", "PORT=x"), Config{}, []string{"PORT"}},
		{"both bad", with("PORT=x"), Config{}, []string{"APP_ENV", "PORT"}},
		{"nothing set", nil, Config{}, []string{"APP_ENV", "POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "AI_HOST: required"}},
		{"postgres empty", []string{"POSTGRES_HOST=", "POSTGRES_USER=", "POSTGRES_PASSWORD=", "POSTGRES_DB="}, Config{},
			[]string{"APP_ENV", "POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"}},
		{"invalid POSTGRES_PORT", with("POSTGRES_PORT=x"), Config{}, []string{"APP_ENV", "POSTGRES_PORT: invalid"}},
		{"zero POSTGRES_PORT", with("POSTGRES_PORT=0"), Config{}, []string{"APP_ENV", "POSTGRES_PORT: invalid"}},
		{"out-of-range POSTGRES_PORT", with("POSTGRES_PORT=65536"), Config{}, []string{"APP_ENV", "POSTGRES_PORT: invalid"}},
		{"missing AI_HOST", append([]string{"APP_ENV=dev"}, pg...), Config{}, []string{"AI_HOST: required"}},
		{"empty AI_HOST", with("APP_ENV=dev", "AI_HOST="), Config{}, []string{"AI_HOST: required"}},
		{"invalid AI_PORT", with("APP_ENV=dev", "AI_PORT=x"), Config{}, []string{"AI_PORT: invalid"}},
		{"zero AI_PORT", with("APP_ENV=dev", "AI_PORT=0"), Config{}, []string{"AI_PORT: invalid"}},
		{"out-of-range AI_PORT", with("APP_ENV=dev", "AI_PORT=65536"), Config{}, []string{"AI_PORT: invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.environ)
			if tt.errNames == nil {
				if err != nil {
					t.Fatalf("Parse: %v", err)
				}
				if got != tt.want {
					t.Fatalf("got %+v, want %+v", got, tt.want)
				}
				return
			}
			if err == nil {
				t.Fatal("Parse: want error, got nil")
			}
			for _, n := range tt.errNames {
				if !strings.Contains(err.Error(), n) {
					t.Errorf("error %q does not name %s", err, n)
				}
			}
			if strings.Contains(tt.name, "AI_PORT") && strings.Contains(err.Error(), "POSTGRES_PORT") {
				t.Errorf("error %q blames POSTGRES_PORT for AI_PORT", err)
			}
			for _, field := range []string{"AppEnv", "Port", "Database", "Host", "User", "Password", "Name"} {
				if strings.Contains(err.Error(), field) {
					t.Errorf("error %q names Go field %s", err, field)
				}
			}
		})
	}
}
