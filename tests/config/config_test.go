package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/config"
)

func TestLoadMySQL(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected config.MySQLConfig
	}{
		{
			name: "defaults",
			env:  map[string]string{},
			expected: config.MySQLConfig{
				Host: "localhost", Port: "3306", User: "root",
				MaxOpenConns: 25, MaxIdleConns: 10, ConnMaxLifetime: 5 * time.Minute,
			},
		},
		{
			name: "env overrides",
			env: map[string]string{
				"MYSQL_HOST": "db.internal", "MYSQL_PORT": "3307", "MYSQL_USER": "svc",
				"MYSQL_PASSWORD": "secret", "MYSQL_DATABASE": "app",
				"MYSQL_MAX_OPEN_CONNS": "50", "MYSQL_MAX_IDLE_CONNS": "5",
				"MYSQL_CONN_MAX_LIFETIME": "1m",
			},
			expected: config.MySQLConfig{
				Host: "db.internal", Port: "3307", User: "svc", Password: "secret", Database: "app",
				MaxOpenConns: 50, MaxIdleConns: 5, ConnMaxLifetime: time.Minute,
			},
		},
		{
			name: "invalid int/duration fall back to default",
			env: map[string]string{
				"MYSQL_MAX_OPEN_CONNS":    "not-a-number",
				"MYSQL_CONN_MAX_LIFETIME": "not-a-duration",
			},
			expected: config.MySQLConfig{
				Host: "localhost", Port: "3306", User: "root",
				MaxOpenConns: 25, MaxIdleConns: 10, ConnMaxLifetime: 5 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			assert.Equal(t, tt.expected, config.LoadMySQL())
		})
	}
}

func TestLoadPostgreSQL(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected config.PostgreSQLConfig
	}{
		{
			name: "defaults",
			env:  map[string]string{},
			expected: config.PostgreSQLConfig{
				Host: "localhost", Port: "5432", User: "postgres", SSLMode: "disable",
				MaxOpenConns: 25, MaxIdleConns: 10, ConnMaxLifetime: 5 * time.Minute,
			},
		},
		{
			name: "env overrides",
			env: map[string]string{
				"POSTGRES_HOST": "pg.internal", "POSTGRES_PORT": "5433", "POSTGRES_USER": "svc",
				"POSTGRES_PASSWORD": "secret", "POSTGRES_DATABASE": "app", "POSTGRES_SSLMODE": "require",
			},
			expected: config.PostgreSQLConfig{
				Host: "pg.internal", Port: "5433", User: "svc", Password: "secret", Database: "app",
				SSLMode: "require", MaxOpenConns: 25, MaxIdleConns: 10, ConnMaxLifetime: 5 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			assert.Equal(t, tt.expected, config.LoadPostgreSQL())
		})
	}
}

func TestLoadCache(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected config.CacheConfig
	}{
		{
			name:     "defaults",
			env:      map[string]string{},
			expected: config.CacheConfig{Host: "localhost", Port: "6379", Prefix: "zus-go"},
		},
		{
			name: "env overrides",
			env: map[string]string{
				"CACHE_HOST": "cache.internal", "CACHE_PORT": "6380",
				"CACHE_PASSWORD": "secret", "CACHE_USER": "svc", "CACHE_PREFIX": "custom",
			},
			expected: config.CacheConfig{
				Host: "cache.internal", Port: "6380", Password: "secret", Username: "svc", Prefix: "custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			assert.Equal(t, tt.expected, config.LoadCache())
		})
	}
}

func TestLoadLogger(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected config.LoggerConfig
	}{
		{
			name:     "defaults",
			env:      map[string]string{},
			expected: config.LoggerConfig{LogLevel: "Debug", ServiceName: "zus-go", HandlerType: "text"},
		},
		{
			name:     "env overrides",
			env:      map[string]string{"LOG_LEVEL": "Error", "LOG_SERVICE_NAME": "svc", "LOG_HANDLER_TYPE": "json"},
			expected: config.LoggerConfig{LogLevel: "Error", ServiceName: "svc", HandlerType: "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			assert.Equal(t, tt.expected, config.LoadLogger())
		})
	}
}

func TestLoadOptional(t *testing.T) {
	type dst struct {
		FeatureFlag bool          `env:"FEATURE_FLAG"`
		MaxQueue    int           `env:"MAX_QUEUE,default=100"`
		CacheTTL    time.Duration `env:"CACHE_TTL,default=5m"`
		Webhook     string        `env:"WEBHOOK_URL,required"`
		Tags        []string      `env:"TAGS"`
		Untagged    string
	}

	tests := []struct {
		name        string
		env         map[string]string
		expected    dst
		expectedErr bool
	}{
		{
			name: "all set",
			env: map[string]string{
				"FEATURE_FLAG": "true", "MAX_QUEUE": "200", "CACHE_TTL": "1m",
				"WEBHOOK_URL": "https://example.com", "TAGS": "a, b,c",
			},
			expected: dst{FeatureFlag: true, MaxQueue: 200, CacheTTL: time.Minute, Webhook: "https://example.com", Tags: []string{"a", "b", "c"}},
		},
		{
			name:        "required missing errors",
			env:         map[string]string{},
			expectedErr: true,
		},
		{
			name:        "invalid bool errors",
			env:         map[string]string{"FEATURE_FLAG": "not-a-bool", "WEBHOOK_URL": "x"},
			expectedErr: true,
		},
		{
			name:     "defaults applied when unset",
			env:      map[string]string{"WEBHOOK_URL": "x", "FEATURE_FLAG": "false"},
			expected: dst{MaxQueue: 100, CacheTTL: 5 * time.Minute, Webhook: "x", Tags: []string{""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			var got dst
			err := config.LoadOptional(&got)
			if tt.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}

	t.Run("non-pointer errors", func(t *testing.T) {
		require.Error(t, config.LoadOptional(dst{}))
	})

	t.Run("nil pointer errors", func(t *testing.T) {
		require.Error(t, config.LoadOptional((*dst)(nil)))
	})
}
