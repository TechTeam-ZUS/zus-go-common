package cache_test

import (
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/TechTeam-ZUS/zus-go-common/cache"
)

func TestHook_ProcessHook_PrefixesKeys(t *testing.T) {
	tests := []struct {
		name         string
		prefix       string
		args         []any
		expectedArgs []any
	}{
		// single-key commands: only args[1] is prefixed
		{name: "get", prefix: "app", args: []any{"get", "k"}, expectedArgs: []any{"get", "app:k"}},
		{name: "set", prefix: "app", args: []any{"set", "k", "v"}, expectedArgs: []any{"set", "app:k", "v"}},
		{name: "setex", prefix: "app", args: []any{"setex", "k", "100", "v"}, expectedArgs: []any{"setex", "app:k", "100", "v"}},
		{name: "setnx", prefix: "app", args: []any{"setnx", "k", "v"}, expectedArgs: []any{"setnx", "app:k", "v"}},
		{name: "expire", prefix: "app", args: []any{"expire", "k", "100"}, expectedArgs: []any{"expire", "app:k", "100"}},
		{name: "ttl", prefix: "app", args: []any{"ttl", "k"}, expectedArgs: []any{"ttl", "app:k"}},
		{name: "pttl", prefix: "app", args: []any{"pttl", "k"}, expectedArgs: []any{"pttl", "app:k"}},
		{name: "persist", prefix: "app", args: []any{"persist", "k"}, expectedArgs: []any{"persist", "app:k"}},
		{name: "incr", prefix: "app", args: []any{"incr", "k"}, expectedArgs: []any{"incr", "app:k"}},
		{name: "incrby", prefix: "app", args: []any{"incrby", "k", "5"}, expectedArgs: []any{"incrby", "app:k", "5"}},
		{name: "decr", prefix: "app", args: []any{"decr", "k"}, expectedArgs: []any{"decr", "app:k"}},
		{name: "decrby", prefix: "app", args: []any{"decrby", "k", "5"}, expectedArgs: []any{"decrby", "app:k", "5"}},
		{name: "hset", prefix: "app", args: []any{"hset", "k", "f", "v"}, expectedArgs: []any{"hset", "app:k", "f", "v"}},
		{name: "hget", prefix: "app", args: []any{"hget", "k", "f"}, expectedArgs: []any{"hget", "app:k", "f"}},
		{name: "hdel", prefix: "app", args: []any{"hdel", "k", "f"}, expectedArgs: []any{"hdel", "app:k", "f"}},
		{name: "hgetall", prefix: "app", args: []any{"hgetall", "k"}, expectedArgs: []any{"hgetall", "app:k"}},
		{name: "hexists", prefix: "app", args: []any{"hexists", "k", "f"}, expectedArgs: []any{"hexists", "app:k", "f"}},
		{name: "hincrby", prefix: "app", args: []any{"hincrby", "k", "f", "5"}, expectedArgs: []any{"hincrby", "app:k", "f", "5"}},
		{name: "hkeys", prefix: "app", args: []any{"hkeys", "k"}, expectedArgs: []any{"hkeys", "app:k"}},
		{name: "hvals", prefix: "app", args: []any{"hvals", "k"}, expectedArgs: []any{"hvals", "app:k"}},
		{name: "hlen", prefix: "app", args: []any{"hlen", "k"}, expectedArgs: []any{"hlen", "app:k"}},
		{name: "lpush", prefix: "app", args: []any{"lpush", "k", "v"}, expectedArgs: []any{"lpush", "app:k", "v"}},
		{name: "rpush", prefix: "app", args: []any{"rpush", "k", "v"}, expectedArgs: []any{"rpush", "app:k", "v"}},
		{name: "lpop", prefix: "app", args: []any{"lpop", "k"}, expectedArgs: []any{"lpop", "app:k"}},
		{name: "rpop", prefix: "app", args: []any{"rpop", "k"}, expectedArgs: []any{"rpop", "app:k"}},
		{name: "llen", prefix: "app", args: []any{"llen", "k"}, expectedArgs: []any{"llen", "app:k"}},
		{name: "sadd", prefix: "app", args: []any{"sadd", "k", "m"}, expectedArgs: []any{"sadd", "app:k", "m"}},
		{name: "srem", prefix: "app", args: []any{"srem", "k", "m"}, expectedArgs: []any{"srem", "app:k", "m"}},
		{name: "smembers", prefix: "app", args: []any{"smembers", "k"}, expectedArgs: []any{"smembers", "app:k"}},
		{name: "scard", prefix: "app", args: []any{"scard", "k"}, expectedArgs: []any{"scard", "app:k"}},
		{name: "zadd", prefix: "app", args: []any{"zadd", "k", "1", "m"}, expectedArgs: []any{"zadd", "app:k", "1", "m"}},
		{name: "zrem", prefix: "app", args: []any{"zrem", "k", "m"}, expectedArgs: []any{"zrem", "app:k", "m"}},
		{name: "zrange", prefix: "app", args: []any{"zrange", "k", "0", "-1"}, expectedArgs: []any{"zrange", "app:k", "0", "-1"}},
		{name: "zscore", prefix: "app", args: []any{"zscore", "k", "m"}, expectedArgs: []any{"zscore", "app:k", "m"}},
		{name: "exists", prefix: "app", args: []any{"exists", "k"}, expectedArgs: []any{"exists", "app:k"}},

		// multi-key commands: every key arg (from index 1) is prefixed
		{name: "del", prefix: "app", args: []any{"del", "k1", "k2"}, expectedArgs: []any{"del", "app:k1", "app:k2"}},
		{name: "unlink", prefix: "app", args: []any{"unlink", "k1", "k2"}, expectedArgs: []any{"unlink", "app:k1", "app:k2"}},
		{name: "mget", prefix: "app", args: []any{"mget", "k1", "k2"}, expectedArgs: []any{"mget", "app:k1", "app:k2"}},

		// pass-through cases
		{name: "unrecognized command passes through", prefix: "app", args: []any{"ping"}, expectedArgs: []any{"ping"}},
		{name: "command with too few args passes through", prefix: "app", args: []any{"get"}, expectedArgs: []any{"get"}},

		// prefix parameter itself varies
		{name: "different prefix, single-key", prefix: "svc", args: []any{"get", "k"}, expectedArgs: []any{"get", "svc:k"}},
		{name: "different prefix, multi-key", prefix: "svc", args: []any{"mget", "k1", "k2"}, expectedArgs: []any{"mget", "svc:k1", "svc:k2"}},
		{name: "empty prefix", prefix: "", args: []any{"get", "k"}, expectedArgs: []any{"get", ":k"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := cache.NewHook(tt.prefix)
			cmd := goredis.NewCmd(context.Background(), tt.args...)

			var calledArgs []any
			next := func(ctx context.Context, c goredis.Cmder) error {
				calledArgs = c.Args()
				return nil
			}

			err := hook.ProcessHook(next)(context.Background(), cmd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedArgs, calledArgs)
		})
	}
}

func TestHook_ProcessPipelineHook_PrefixesAllCommands(t *testing.T) {
	tests := []struct {
		name         string
		prefix       string
		args         [][]any
		expectedArgs [][]any
	}{
		{
			name:         "prefixes every command in the pipeline",
			prefix:       "app",
			args:         [][]any{{"get", "a"}, {"get", "b"}},
			expectedArgs: [][]any{{"get", "app:a"}, {"get", "app:b"}},
		},
		{
			name:         "mixes recognized and pass-through commands",
			prefix:       "app",
			args:         [][]any{{"get", "a"}, {"ping"}},
			expectedArgs: [][]any{{"get", "app:a"}, {"ping"}},
		},
		{
			name:         "different prefix",
			prefix:       "svc",
			args:         [][]any{{"get", "a"}, {"mget", "b", "c"}},
			expectedArgs: [][]any{{"get", "svc:a"}, {"mget", "svc:b", "svc:c"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := cache.NewHook(tt.prefix)

			cmds := make([]goredis.Cmder, len(tt.args))
			for i, args := range tt.args {
				cmds[i] = goredis.NewCmd(context.Background(), args...)
			}

			var calledCmds []goredis.Cmder
			next := func(ctx context.Context, cmds []goredis.Cmder) error {
				calledCmds = cmds
				return nil
			}

			err := hook.ProcessPipelineHook(next)(context.Background(), cmds)
			assert.NoError(t, err)

			for i, expected := range tt.expectedArgs {
				assert.Equal(t, expected, calledCmds[i].Args())
			}
		})
	}
}
