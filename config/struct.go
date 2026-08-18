package config

import "time"

// type OptionalConfig struct {
// 	notificationBot string `env:"NOTIFICATION_BOT"`
// }

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
	RetryCount      int
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
	RetryCount      int
}

// Cache Config
type CacheConfig struct {
	Host       string
	Port       string
	Password   string
	Username   string
	Prefix     string
	RetryCount int
}

// Logger Config
type LoggerConfig struct {
	LogLevel    string
	ServiceName string
	HandlerType string
}
