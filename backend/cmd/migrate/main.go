// Command migrate applies or rolls back database schema migrations.
//
// It intentionally reads only DATABASE_URL from the environment (not the full
// application config) so migrations stay runnable without API-only settings.
//
// Usage:
//
//	migrate up        # apply all pending migrations (default)
//	migrate down      # roll back the most recent migration
//	migrate version   # print current schema version
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goacademy/backend/db"
	"github.com/goacademy/backend/internal/platform/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	mg, err := postgres.NewMigrator(dsn, db.MigrationsFS, db.MigrationsDir)
	if err != nil {
		return err
	}
	defer mg.Close()

	switch command {
	case "up":
		if err := mg.Up(); err != nil {
			return err
		}
		v, dirty, _ := mg.Version()
		fmt.Printf("migrations applied; version=%d dirty=%v\n", v, dirty)
	case "down":
		if err := mg.Steps(-1); err != nil {
			return err
		}
		v, dirty, _ := mg.Version()
		fmt.Printf("rolled back one migration; version=%d dirty=%v\n", v, dirty)
	case "version":
		v, dirty, err := mg.Version()
		if err != nil {
			return err
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	default:
		return fmt.Errorf("unknown command %q (use: up | down | version)", command)
	}
	return nil
}
