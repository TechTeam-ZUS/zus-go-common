package config

import "time"

// MySQL Config
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

// Postgre Config
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

// Cache Config
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	Username string
	Prefix   string
}

// Logger Config
type LoggerConfig struct {
	LogLevel    string
	ServiceName string
	HandlerType string
}
