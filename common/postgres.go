package common

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupPostgreSQLConnection opens a PostgreSQL connection using the provided config
// and configures the connection pool.
func SetupPostgreSQLConnection(cfg PostgreSQLConfig) (*sql.DB, error) {
	if cfg.Database == "" {
		return nil, fmt.Errorf("postgres database name is required")
	}

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

// SetupPostgreSQLConnectionFromEnv loads PostgreSQL settings from environment variables
// and opens a connection.
func SetupPostgreSQLConnectionFromEnv() (*sql.DB, error) {
	return SetupPostgreSQLConnection(PostgreSQLConfigFromEnv())
}
