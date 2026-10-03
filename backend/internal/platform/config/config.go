// Package config parses backend's settings from the environment.
package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/database"
)

// AppEnv is the deployment environment; its own type so other binaries reuse it.
type AppEnv string

const (
	Dev        AppEnv = "dev"
	Staging    AppEnv = "staging"
	Production AppEnv = "production"
)

func (e *AppEnv) UnmarshalText(b []byte) error {
	switch v := AppEnv(b); v {
	case Dev, Staging, Production:
		*e = v
		return nil
	}
	return errors.New("must be dev, staging or production")
}

type Config struct {
	AppEnv   AppEnv `env:"APP_ENV,required,notEmpty"`
	Port     int    `env:"PORT" envDefault:"8080"`
	Database database.Config
	AI       AI
}

// AI is declared here, not in platform/ai, because platform/ai will import logger, which imports config.
type AI struct {
	Host string `env:"AI_HOST,required,notEmpty"`
	Port AIPort `env:"AI_PORT" envDefault:"50051"`
}

// AIPort is not database.Port: envKey matches by field name and type, so sharing it would report AI_PORT as POSTGRES_PORT.
type AIPort uint16

func (p *AIPort) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 10, 16)
	if err != nil || n == 0 {
		return errors.New("must be a port from 1 to 65535")
	}
	*p = AIPort(n)
	return nil
}

// Parse reads environ (KEY=value pairs) and reports every bad variable in one error.
func Parse(environ []string) (Config, error) { return ParseAs[Config](environ) }

// ParseAs is Parse for a binary that needs only some of the settings.
func ParseAs[T any](environ []string) (T, error) {
	var zero T
	cfg, err := env.ParseAsWithOptions[T](env.Options{Environment: env.ToMap(environ)})
	if err == nil {
		return cfg, nil
	}
	var agg env.AggregateError
	if !errors.As(err, &agg) {
		return zero, fmt.Errorf("config: %w", err)
	}
	msgs := make([]string, 0, len(agg.Errors))
	for _, e := range agg.Errors {
		msgs = append(msgs, describe(reflect.TypeFor[T](), e))
	}
	return zero, fmt.Errorf("config: %s", strings.Join(msgs, "; "))
}

// describe names the variable; env's ParseError names only the Go field.
func describe(t reflect.Type, err error) string {
	var unset env.VarIsNotSetError
	if errors.As(err, &unset) {
		return unset.Key + ": required"
	}
	var empty env.EmptyVarError
	if errors.As(err, &empty) {
		return empty.Key + ": required"
	}
	var parse env.ParseError
	if errors.As(err, &parse) {
		if key, ok := envKey(t, parse); ok {
			return fmt.Sprintf("%s: invalid value: %v", key, parse.Err)
		}
	}
	return err.Error()
}

// envKey searches nested sub-configs too, matching the type as well as the name
// because ParseError carries only the field name and two sub-configs may share one.
func envKey(t reflect.Type, parse env.ParseError) (string, bool) {
	for f := range t.Fields() {
		if f.Type.Kind() == reflect.Struct {
			if key, ok := envKey(f.Type, parse); ok {
				return key, true
			}
			continue
		}
		if f.Name == parse.Name && f.Type == parse.Type {
			key, _, _ := strings.Cut(f.Tag.Get("env"), ",")
			return key, true
		}
	}
	return "", false
}
