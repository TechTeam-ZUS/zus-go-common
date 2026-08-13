package mysql

import (
	"database/sql"
	"fmt"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	_ "github.com/go-sql-driver/mysql"
)

// SetupMySQLConnection opens a MySQL connection using the provided config
// and configures the connection pool.
func Init() (*sql.DB, error) {
	cfg := config.LoadMySQL()

	if cfg.Database == "" {
		return nil, fmt.Errorf("mysql database name is required")
	}

	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}
