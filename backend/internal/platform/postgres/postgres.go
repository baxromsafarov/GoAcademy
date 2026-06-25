// Package postgres provides the PostgreSQL connection pool and the schema
// migration runner used by the GoAcademy backend.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect options. Defaults tolerate a database that is still starting up
// (e.g. inside docker compose) by retrying the initial ping.
const (
	connectAttempts = 10
	connectBackoff  = time.Second
	pingTimeout     = 3 * time.Second
)

// Connect creates a pgx connection pool and verifies connectivity, retrying the
// initial ping with a fixed backoff. It returns an error if the database cannot
// be reached within connectAttempts or if ctx is cancelled.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	var pingErr error
	for attempt := 1; attempt <= connectAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
		pingErr = pool.Ping(pingCtx)
		cancel()
		if pingErr == nil {
			return pool, nil
		}
		if attempt == connectAttempts {
			break
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(connectBackoff):
		}
	}

	pool.Close()
	return nil, fmt.Errorf("database unreachable after %d attempts: %w", connectAttempts, pingErr)
}
