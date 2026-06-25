// Package db embeds the SQL migration files so the migration runner works
// regardless of the process working directory (useful in containers).
package db

import "embed"

// MigrationsFS holds the versioned migration files under migrations/.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS

// MigrationsDir is the subdirectory within MigrationsFS holding the *.sql files.
const MigrationsDir = "migrations"
