// Package buildinfo holds values stamped at build time with -ldflags -X.
package buildinfo

var (
	Version = "dev"
	// Commit is logged at startup only, never returned to clients.
	Commit = "dev"
)
