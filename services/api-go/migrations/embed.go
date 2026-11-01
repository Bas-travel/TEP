// Package migrations embeds the SQL migration files (golang-migrate naming: NNNN_name.up.sql / .down.sql).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
