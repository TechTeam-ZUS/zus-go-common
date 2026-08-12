# zus-go-common

Shared Go package for ZUS services. Centralizes database and cache connection setup so consuming projects inherit consistent driver versions from a single module.

## Installation

```bash
go get github.com/TechTeam-ZUS/zus-go-common
```

## Setup

Copy the example env file and fill in your credentials:

```bash
cp .env.example .env
```

## Usage

```go
package main

import (
    "log"

    "github.com/TechTeam-ZUS/zus-go-common/common"
)

func main() {
    if err := common.LoadEnv(); err != nil {
        log.Fatal(err)
    }

    mysqlDB, err := common.SetupMySQLConnectionFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer mysqlDB.Close()

    pgDB, err := common.SetupPostgreSQLConnectionFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer pgDB.Close()

    redisClient, err := common.SetupRedisConnectionFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer redisClient.Close()
}
```

You can also pass explicit config instead of reading from common environment file:

```go
db, err := common.SetupMySQLConnection(common.MySQLConfig{
    Host:     "localhost",
    Port:     "3306",
    User:     "root",
    Password: "secret",
    Database: "myapp",
})
```

## Environment Variables

See [.env.example](.env.example) for all supported variables.

| Service    | Required Variables                          |
|------------|---------------------------------------------|
| MySQL      | `MYSQL_DATABASE`, `MYSQL_PASSWORD`          |
| PostgreSQL | `POSTGRES_DATABASE`, `POSTGRES_PASSWORD`    |
| Redis      | None (defaults to `localhost:6379`, DB `0`) |

## API

| Function | Description |
|----------|-------------|
| `LoadEnv()` | Load variables from `.env` |
| `SetupMySQLConnection(cfg)` | Open and ping a MySQL pool |
| `SetupMySQLConnectionFromEnv()` | MySQL setup using env vars |
| `SetupPostgreSQLConnection(cfg)` | Open and ping a PostgreSQL pool |
| `SetupPostgreSQLConnectionFromEnv()` | PostgreSQL setup using env vars |
| `SetupRedisConnection(cfg)` | Create and ping a Redis client |
| `SetupRedisConnectionFromEnv()` | Redis setup using env vars |
