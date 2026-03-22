package storage

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool   *pgxpool.Pool
	poolMu sync.Once
)

// Pool returns a shared Postgres connection pool (creates it on first call).
func Pool() *pgxpool.Pool {
	poolMu.Do(func() {
		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_DB"))

		var err error
		pool, err = pgxpool.New(context.Background(), dbURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}
	})
	return pool
}

// HealthCheck pings the database and returns nil on success.
func HealthCheck(ctx context.Context) error {
	return pool.Ping(ctx)
}
