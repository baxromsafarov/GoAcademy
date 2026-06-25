package postgres

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // registers the postgres:// driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrator applies versioned SQL migrations (read from an embedded fs.FS) to a
// PostgreSQL database. Up/down pairs are tracked in the schema_migrations table.
//
// Note: golang-migrate's postgres driver uses lib/pq internally; the application
// runtime uses pgx. Both speak the same postgres:// DSN (see decision D-011).
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator builds a Migrator reading *.sql files from fsys under dir and
// applying them to the database at dbURL.
func NewMigrator(dbURL string, fsys fs.FS, dir string) (*Migrator, error) {
	src, err := iofs.New(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("open migration source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}
	return &Migrator{m: m}, nil
}

// Up applies all pending migrations. Being already up to date is not an error.
func (mg *Migrator) Up() error {
	if err := mg.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Steps moves n migrations forward (n>0) or back (n<0).
func (mg *Migrator) Steps(n int) error {
	if err := mg.m.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Version returns the current schema version and whether it is dirty (a previous
// migration failed mid-way). Version 0 means no migrations have been applied.
func (mg *Migrator) Version() (version uint, dirty bool, err error) {
	v, d, err := mg.m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return v, d, err
}

// Close releases the migrator's source and database handles.
func (mg *Migrator) Close() error {
	srcErr, dbErr := mg.m.Close()
	return errors.Join(srcErr, dbErr)
}
