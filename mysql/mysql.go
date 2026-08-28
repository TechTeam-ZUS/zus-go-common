package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	"github.com/TechTeam-ZUS/zus-go-common/retry"
	_ "github.com/go-sql-driver/mysql"
)

// SetupMySQLConnection opens a MySQL connection using the provided config and configures the connection pool.
// Retries cfg.RetryCount times before giving up; exhausting retries is fatal.
func Init() (*sql.DB, error) {
	cfg := config.LoadMySQL()

	if cfg.Database == "" {
		return nil, fmt.Errorf("mysql database name is required")
	}

	var db *sql.DB
	err := retry.Do(cfg.RetryCount, retry.RetryDelay, func() error {
		conn, err := sql.Open("mysql", dsn(cfg))
		if err != nil {
			return err
		}

		conn.SetMaxOpenConns(cfg.MaxOpenConns)
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
		conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)
		//ping timeout for 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err = conn.PingContext(ctx); err != nil {
			_ = conn.Close()
			return err
		}

		db = conn
		return nil
	}, "MySQL Connection")

	if err != nil {
		return nil, fmt.Errorf("Failed to connect MySQL: %w", err)
	}

	return db, nil
}

func dsn(cfg config.MySQLConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}
