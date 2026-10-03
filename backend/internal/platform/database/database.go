// Package database owns backend's PostgreSQL settings and its sqlx pool.
package database

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the pgx driver
	"github.com/jmoiron/sqlx"
)

// Config holds the database settings; it embeds the password, so never log it.
type Config struct {
	Host     string `env:"POSTGRES_HOST,required,notEmpty"`
	Port     Port   `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER,required,notEmpty"`
	Password string `env:"POSTGRES_PASSWORD,required,notEmpty"`
	Name     string `env:"POSTGRES_DB,required,notEmpty"`
}

// Port rejects 0, which parses as a uint16 but is a port nothing listens on.
type Port uint16

func (p *Port) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 10, 16)
	if err != nil || n == 0 {
		return errors.New("must be a port from 1 to 65535")
	}
	*p = Port(n)
	return nil
}

// connString escapes each part, so credentials may hold URL-special characters.
func (c Config) connString() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port))),
		Path:   "/" + c.Name,
	}
	return u.String()
}

// Open returns a lazy pool: nothing dials until first use, so boot never needs PostgreSQL.
func Open(cfg Config) (*sqlx.DB, error) {
	// "pgx", not "pgx/v5": only "pgx" is in sqlx's bindvar table, giving $n placeholders.
	db, err := sqlx.Open("pgx", cfg.connString())
	if err != nil {
		return nil, fmt.Errorf("opening the database pool: %w", err)
	}
	return db, nil
}
