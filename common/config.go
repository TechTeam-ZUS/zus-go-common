package common

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// MySQLConfig holds MySQL connection settings.
type MySQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// PostgreSQLConfig holds PostgreSQL connection settings.
type PostgreSQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// LoadEnv loads environment variables from a .env file.
// If paths are omitted, it loads ".env" from the current working directory.
func LoadEnv(paths ...string) error {
	if len(paths) == 0 {
		return godotenv.Load()
	}

	return godotenv.Load(paths...)
}

// MySQLConfigFromEnv reads MySQL settings from environment variables.
func MySQLConfigFromEnv() MySQLConfig {
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
func PostgreSQLConfigFromEnv() PostgreSQLConfig {
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
func RedisConfigFromEnv() RedisConfig {
	return RedisConfig{
		Host:     envOrDefault("REDIS_HOST", "localhost"),
		Port:     envOrDefault("REDIS_PORT", "6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       envIntOrDefault("REDIS_DB", 0),
	}
}

// Addr returns the Redis host:port address.
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
