package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	"github.com/TechTeam-ZUS/zus-go-common/logger"
	"github.com/TechTeam-ZUS/zus-go-common/retry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupPostgreSQLConnection opens a PostgreSQL connection using the provided config and configures the connection pool.
// Retries cfg.RetryCount times before giving up; exhausting retries is fatal.
func Init() (*sql.DB, error) {
	cfg := config.LoadPostgreSQL()

	if cfg.Database == "" {
		return nil, fmt.Errorf("postgres database name is required")
	}

	var db *sql.DB
	err := retry.Do(cfg.RetryCount, retry.RetryDelay, func() error {
		db, err := sql.Open("pgx", dsn(cfg))
		if err != nil {
			return err
		}

		db.SetMaxOpenConns(cfg.MaxOpenConns)
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if err != nil {
			logger.Fatal("postgres: failed to connect after retries", "error", err.Error())
		}

		//ping timeout for 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			_ = db.Close()
			return err
		}

		return nil
	}, "Postgres Connection")

	if err != nil {
		return nil, fmt.Errorf("Failed to connect Postgres: %w", err)
	}

	return db, nil
}

func dsn(cfg config.PostgreSQLConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
}
