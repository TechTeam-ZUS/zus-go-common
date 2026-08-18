# zus-go-common

Shared Go package for ZUS services. Centralizes config loading, database, cache, and logging setup so consuming projects inherit consistent driver versions and conventions from a single module.

## Installation

```bash
go get github.com/TechTeam-ZUS/zus-go-common
```

## Packages

| Package  | Purpose                                              |
|----------|-------------------------------------------------------|
| `config` | Loads `.env` and env-backed config for every package |
| `mysql`  | MySQL connection pool setup                          |
| `postgres` | PostgreSQL connection pool setup                   |
| `cache`  | Cache client setup (Redis/Valkey-protocol), with automatic key prefixing |
| `logger` | `slog`-based structured logger setup                 |

Each of `mysql`, `postgres`, `cache`, and `logger` reads its own configuration from environment variables internally — call `config.Load()` once at startup, then call `Init()` on whichever packages you need.

## Usage

```go
package main

import (
    "log"

    "github.com/TechTeam-ZUS/zus-go-common/cache"
    "github.com/TechTeam-ZUS/zus-go-common/config"
    "github.com/TechTeam-ZUS/zus-go-common/logger"
    "github.com/TechTeam-ZUS/zus-go-common/mysql"
    "github.com/TechTeam-ZUS/zus-go-common/postgres"
)

func main() {
    // Loads .env into the process environment.
    if err := config.Load(nil); err != nil {
        log.Fatal(err)
    }

    log := logger.Init()

    mysqlDB, err := mysql.Init()
    if err != nil {
        log.Error("mysql init failed", "error", err)
        return
    }
    defer mysqlDB.Close()

    pgDB, err := postgres.Init()
    if err != nil {
        log.Error("postgres init failed", "error", err)
        return
    }
    defer pgDB.Close()

    cacheInstance, err := cache.Init()
    if err != nil {
        log.Error("cache init failed", "error", err)
        return
    }
    defer cacheInstance.Close()
}
```

`mysql.Init()` and `postgres.Init()` return a standard `*sql.DB`, already pinged and pool-configured. `cache.Init()` returns a `*cache.CacheInstance` wrapping a `*redis.Client` from the go-redis driver — compatible with both Redis and Valkey servers (see [Cache key prefixing](#cache-key-prefixing) below). `logger.Init()` returns a `*slog.Logger`.

Connection pool settings, credentials, and other tuning are read from environment variables — see [Environment Variables](#environment-variables).

## Custom / optional config

Beyond the built-in MySQL, PostgreSQL, cache, and logger config, `config.Load` can also populate a consumer-defined struct from environment variables using an `env` struct tag:

```go
type MyConfig struct {
    FeatureFlagX   bool          `env:"FEATURE_FLAG_X"`
    MaxQueueSize   int           `env:"MAX_QUEUE_SIZE,default=100"`
    CacheTTL       time.Duration `env:"CACHE_TTL,default=5m"`
    PaymentWebhook string        `env:"PAYMENT_WEBHOOK_URL,required"`
}

var cfg MyConfig
if err := config.Load(&cfg); err != nil {
    log.Fatal(err)
}
```

Tag options (comma-separated after the key):

| Option           | Behavior if unset            |
|------------------|-------------------------------|
| `env:"KEY"`              | Field keeps its zero value |
| `env:"KEY,required"`     | `Load` returns an error    |
| `env:"KEY,default=value"`| Field is set to `value`    |

Supported field types: `string`, `bool`, all integer kinds, `float32`/`float64`, `time.Duration`, and `[]string` (comma-separated values). Fields without an `env` tag are left untouched. Pass `nil` to `config.Load` if you only need the built-in configs and have no custom struct.

## Cache key prefixing

`cache.Init()` attaches a hook that automatically prefixes every key with `CACHE_PREFIX + ":"` for common single- and multi-key commands (`GET`, `SET`, `HSET`, `DEL`, `MGET`, etc.). Callers don't need to build the prefixed key themselves — just use normal key names and the client namespaces them transparently. Unrecognized commands (`SCAN`, `PING`, `INFO`, etc.) pass through unchanged.

## Environment Variables

### MySQL

| Variable | Default | Required |
|----------|---------|----------|
| `MYSQL_HOST` | `localhost` | No |
| `MYSQL_PORT` | `3306` | No |
| `MYSQL_USER` | `root` | No |
| `MYSQL_PASSWORD` | — | Yes |
| `MYSQL_DATABASE` | — | Yes |
| `MYSQL_MAX_OPEN_CONNS` | `25` | No |
| `MYSQL_MAX_IDLE_CONNS` | `10` | No |
| `MYSQL_CONN_MAX_LIFETIME` | `5m` | No |

### PostgreSQL

| Variable | Default | Required |
|----------|---------|----------|
| `POSTGRES_HOST` | `localhost` | No |
| `POSTGRES_PORT` | `5432` | No |
| `POSTGRES_USER` | `postgres` | No |
| `POSTGRES_PASSWORD` | — | Yes |
| `POSTGRES_DATABASE` | — | Yes |
| `POSTGRES_SSLMODE` | `disable` | No |
| `POSTGRES_MAX_OPEN_CONNS` | `25` | No |
| `POSTGRES_MAX_IDLE_CONNS` | `10` | No |
| `POSTGRES_CONN_MAX_LIFETIME` | `5m` | No |

### Cache

| Variable | Default | Required |
|----------|---------|----------|
| `CACHE_HOST` | `localhost` | No |
| `CACHE_PORT` | `6379` | No |
| `CACHE_USER` | — | No |
| `CACHE_PASSWORD` | — | No |
| `CACHE_PREFIX` | `zus-go` | No |

### Logger

| Variable | Default | Required |
|----------|---------|----------|
| `LOG_LEVEL` | `Debug` | No |
| `LOG_SERVICE_NAME` | `zus-go` | No |
| `LOG_HANDLER_TYPE` | `text` | No |

## API

### config

| Function | Description |
|----------|-------------|
| `Load(dst any, paths ...string) error` | Loads `.env` (or the given paths). If `dst` is non-nil, also fills it via `LoadOptional`. |
| `LoadOptional(dst any) error` | Fills a consumer-defined struct from env vars using `env` tags. Called internally by `Load`, but usable standalone. |
| `LoadMySQL() MySQLConfig` | Reads MySQL settings from env vars. |
| `LoadPostgreSQL() PostgreSQLConfig` | Reads PostgreSQL settings from env vars. |
| `LoadCache() CacheConfig` | Reads cache settings from env vars. |
| `LoadLogger() LoggerConfig` | Reads logger settings from env vars. |

### mysql

| Function | Description |
|----------|-------------|
| `Init() (*sql.DB, error)` | Reads MySQL config from env, opens and pings a connection pool. |

### postgres

| Function | Description |
|----------|-------------|
| `Init() (*sql.DB, error)` | Reads PostgreSQL config from env, opens and pings a connection pool. |

### cache

| Function | Description |
|----------|-------------|
| `Init() (*CacheInstance, error)` | Reads cache config from env, creates and pings a client with the key-prefixing hook attached. |
| `(CacheInstance) Close() error` | Closes the underlying client. |

### logger

| Function | Description |
|----------|-------------|
| `Init() *slog.Logger` | Reads logger config from env and returns a configured `slog.Logger`. |
