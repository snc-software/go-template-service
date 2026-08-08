// Package migrations embeds the goose migration files so they can be applied
// from Go without depending on the working directory.
package migrations

import "embed"

// FS holds every migration in this directory.
//
//go:embed *.sql
var FS embed.FS
