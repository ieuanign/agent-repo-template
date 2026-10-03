// Package migrations embeds backend's SQL migrations.
package migrations

import (
	"embed"
	"io/fs"
)

// "*" also matches this file, so the pattern compiles while no .sql file exists; iofs skips it.
//
//go:embed *
var files embed.FS

var FS fs.FS = files
