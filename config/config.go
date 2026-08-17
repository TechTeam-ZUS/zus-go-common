package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file.
// If paths are omitted, it loads ".env" from the current working directory.
// Fills optional config if any
func Load(dst any, paths ...string) error {
	var err error
	if len(paths) == 0 {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(paths...)
	}

	if err != nil {
		return err
	}

	if dst == nil {
		return nil
	}

	return LoadOptional(dst)
}

// MySQLConfigFromEnv reads MySQL settings from environment variables.
func LoadMySQL() MySQLConfig {
	return MySQLConfig{
		Host:            envOrDefault("MYSQL_HOST", "localhost"),
		Port:            envOrDefault("MYSQL_PORT", "3306"),
		User:            envOrDefault("MYSQL_USER", "root"),
		Password:        os.Getenv("MYSQL_PASSWORD"),
		Database:        os.Getenv("MYSQL_DATABASE"),
		MaxOpenConns:    envIntOrDefault("MYSQL_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    envIntOrDefault("MYSQL_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: envDurationOrDefault("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// DSN returns a MySQL data source name.
func (c MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// PostgreSQLConfigFromEnv reads PostgreSQL settings from environment variables.
func LoadPostgreSQL() PostgreSQLConfig {
	return PostgreSQLConfig{
		Host:            envOrDefault("POSTGRES_HOST", "localhost"),
		Port:            envOrDefault("POSTGRES_PORT", "5432"),
		User:            envOrDefault("POSTGRES_USER", "postgres"),
		Password:        os.Getenv("POSTGRES_PASSWORD"),
		Database:        os.Getenv("POSTGRES_DATABASE"),
		SSLMode:         envOrDefault("POSTGRES_SSLMODE", "disable"),
		MaxOpenConns:    envIntOrDefault("POSTGRES_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    envIntOrDefault("POSTGRES_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: envDurationOrDefault("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// DSN returns a PostgreSQL connection string.
func (c PostgreSQLConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

// RedisConfigFromEnv reads Redis settings from environment variables.
func LoadRedis() RedisConfig {
	return RedisConfig{
		Host:     envOrDefault("REDIS_HOST", "localhost"),
		Port:     envOrDefault("REDIS_PORT", "6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Username: os.Getenv("REDIS_USER"),
		Prefix:   envOrDefault("REDIS_PREFIX", "zus-go"),
	}
}

// Addr returns the Redis host:port address.
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func LoadLogger() LoggerConfig {
	return LoggerConfig{
		LogLevel:    envOrDefault("LOG_LEVEL", "Debug"),
		ServiceName: envOrDefault("LOG_SERVICE_NAME", "zus-go"),
		HandlerType: envOrDefault("LOG_HANDLER_TYPE", "text"),
	}
}
