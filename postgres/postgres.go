package postgres

import (
	"database/sql"
	"fmt"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupPostgreSQLConnection opens a PostgreSQL connection using the provided config
// and configures the connection pool.
func Init() (*sql.DB, error) {
	cfg := config.LoadPostgreSQL()

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
