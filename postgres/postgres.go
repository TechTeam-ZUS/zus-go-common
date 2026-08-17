package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupPostgreSQLConnection opens a PostgreSQL connection using the provided config and configures the connection pool.
func Init() (*sql.DB, error) {
	cfg := config.LoadPostgreSQL()

	if cfg.Database == "" {
		return nil, fmt.Errorf("postgres database name is required")
	}

	db, err := sql.Open("pgx", dsn(cfg))
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	//ping timeout for 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

func dsn(cfg config.PostgreSQLConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
}
