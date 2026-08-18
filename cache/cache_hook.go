// Package cache provides a redis.Hook that transparently prefixes every key
// with a fixed namespace (e.g. "myapp:") so callers never need to build the
// prefixed key themselves.
package cache

import (
	"context"
	"strings"

	redis "github.com/redis/go-redis/v9"
)

var isSingleKeyCmdMap = map[string]bool{
	"get": true, "set": true, "setex": true, "setnx": true,
	"expire": true, "ttl": true, "pttl": true, "persist": true,
	"incr": true, "incrby": true, "decr": true, "decrby": true,
	"hset": true, "hget": true, "hdel": true, "hgetall": true,
	"hexists": true, "hincrby": true, "hkeys": true, "hvals": true, "hlen": true,
	"lpush": true, "rpush": true, "lpop": true, "rpop": true, "llen": true,
	"sadd": true, "srem": true, "smembers": true, "scard": true,
	"zadd": true, "zrem": true, "zrange": true, "zscore": true,
	"del": false, "unlink": false, "exists": true, "mget": false,
}

// Hook prefixes cache keys with Prefix + ":" for every command it recognizes.
// Unrecognized commands (SCAN, PING, INFO, etc.) pass through untouched.
type Hook struct {
	prefix string
}

// New returns a Hook that namespaces every key under prefix + ":".
func NewHook(prefix string) *Hook {
	return &Hook{prefix: prefix}
}

func (h *Hook) addPrefix(key string) string {
	return h.prefix + ":" + key
}

func (h *Hook) rewrite(cmd redis.Cmder) {
	args := cmd.Args()
	if len(args) < 2 {
		return
	}

	name := strings.ToLower(cmd.Name())
	isSingleKeyCmd, ok := isSingleKeyCmdMap[name]
	if !ok {
		return
	}

	// rewrite key
	if isSingleKeyCmd {
		if key, ok := args[1].(string); ok {
			args[1] = h.addPrefix(key)
		}
	} else {
		// rewrite all the keys in multi key commands
		for i := 1; i < len(args); i++ {
			if key, ok := args[i].(string); ok {
				args[i] = h.addPrefix(key)
			}
		}
	}
}

// implementations of redis.Hook interfaces to insert key prefix
func (h *Hook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *Hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		h.rewrite(cmd)
		return next(ctx, cmd)
	}
}

func (h *Hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			h.rewrite(cmd)
		}
		return next(ctx, cmds)
	}
}
