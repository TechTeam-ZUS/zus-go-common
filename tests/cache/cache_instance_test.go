package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TechTeam-ZUS/zus-go-common/cache"
)

const testPrefix = "app"

// newTestInstance starts an in-memory Redis-protocol server and returns a
// real CacheInstance from cache.Init(), so tests exercise the actual client
// + prefixing hook end to end rather than just the hook in isolation.
func newTestInstance(t *testing.T) (*cache.CacheInstance, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)

	t.Setenv("CACHE_HOST", mr.Host())
	t.Setenv("CACHE_PORT", mr.Port())
	t.Setenv("CACHE_PREFIX", testPrefix)
	t.Setenv("CACHE_PASSWORD", "")
	t.Setenv("CACHE_USER", "")

	instance, err := cache.Init()
	require.NoError(t, err)
	t.Cleanup(func() { instance.Close() })

	return instance, mr
}

// assertPrefixed checks the key exists under its prefixed name in the
// underlying store, and not under its raw (unprefixed) name.
func assertPrefixed(t *testing.T, mr *miniredis.Miniredis, key string) {
	t.Helper()
	assert.True(t, mr.Exists(testPrefix+":"+key), "expected prefixed key %q to exist", testPrefix+":"+key)
	assert.False(t, mr.Exists(key), "expected raw key %q to NOT exist", key)
}

type instanceCmdCase struct {
	name string
	run  func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis)
}

func runInstanceCmdCases(t *testing.T, tests []instanceCmdCase) {
	instance, mr := newTestInstance(t)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, ctx, instance.Client, mr)
		})
	}
}

func TestCacheInstance_StringCommands(t *testing.T) {
	runInstanceCmdCases(t, []instanceCmdCase{
		{
			name: "SET and GET",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "str1", "v1", 0).Err())
				assertPrefixed(t, mr, "str1")
				got, err := c.Get(ctx, "str1").Result()
				require.NoError(t, err)
				assert.Equal(t, "v1", got)
			},
		},
		{
			name: "SETNX",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				ok, err := c.SetNX(ctx, "str2", "v2", 0).Result()
				require.NoError(t, err)
				assert.True(t, ok)
				assertPrefixed(t, mr, "str2")
			},
		},
		{
			name: "SETEX",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.SetEx(ctx, "str3", "v3", time.Minute).Err())
				assertPrefixed(t, mr, "str3")
			},
		},
		{
			name: "INCR and INCRBY",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "counter", "10", 0).Err())
				n, err := c.Incr(ctx, "counter").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 11, n)
				n, err = c.IncrBy(ctx, "counter", 5).Result()
				require.NoError(t, err)
				assert.EqualValues(t, 16, n)
				assertPrefixed(t, mr, "counter")
			},
		},
		{
			name: "DECR and DECRBY",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "counter2", "10", 0).Err())
				n, err := c.Decr(ctx, "counter2").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 9, n)
				n, err = c.DecrBy(ctx, "counter2", 4).Result()
				require.NoError(t, err)
				assert.EqualValues(t, 5, n)
				assertPrefixed(t, mr, "counter2")
			},
		},
		{
			name: "EXPIRE, TTL and PTTL",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "expiring", "v", 0).Err())
				ok, err := c.Expire(ctx, "expiring", time.Minute).Result()
				require.NoError(t, err)
				assert.True(t, ok)
				ttl, err := c.TTL(ctx, "expiring").Result()
				require.NoError(t, err)
				assert.Positive(t, ttl)
				pttl, err := c.PTTL(ctx, "expiring").Result()
				require.NoError(t, err)
				assert.Positive(t, pttl)
				assertPrefixed(t, mr, "expiring")
			},
		},
		{
			name: "PERSIST",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "persisting", "v", time.Minute).Err())
				ok, err := c.Persist(ctx, "persisting").Result()
				require.NoError(t, err)
				assert.True(t, ok)
				assertPrefixed(t, mr, "persisting")
			},
		},
		{
			name: "EXISTS",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "exists1", "v", 0).Err())
				n, err := c.Exists(ctx, "exists1").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 1, n)
				assertPrefixed(t, mr, "exists1")
			},
		},
	})
}

func TestCacheInstance_MultiKeyCommands(t *testing.T) {
	runInstanceCmdCases(t, []instanceCmdCase{
		{
			name: "MGET",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "mk1", "v1", 0).Err())
				require.NoError(t, c.Set(ctx, "mk2", "v2", 0).Err())
				vals, err := c.MGet(ctx, "mk1", "mk2").Result()
				require.NoError(t, err)
				assert.Equal(t, []any{"v1", "v2"}, vals)
				assertPrefixed(t, mr, "mk1")
				assertPrefixed(t, mr, "mk2")
			},
		},
		{
			name: "DEL",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "del1", "v", 0).Err())
				require.NoError(t, c.Set(ctx, "del2", "v", 0).Err())
				n, err := c.Del(ctx, "del1", "del2").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 2, n)
				assert.False(t, mr.Exists(testPrefix+":del1"))
				assert.False(t, mr.Exists(testPrefix+":del2"))
			},
		},
		{
			name: "UNLINK",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.Set(ctx, "unlink1", "v", 0).Err())
				n, err := c.Unlink(ctx, "unlink1").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 1, n)
				assert.False(t, mr.Exists(testPrefix+":unlink1"))
			},
		},
	})
}

func TestCacheInstance_HashCommands(t *testing.T) {
	runInstanceCmdCases(t, []instanceCmdCase{
		{
			name: "HSET and HGET",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash1", "f1", "v1").Err())
				assertPrefixed(t, mr, "hash1")
				v, err := c.HGet(ctx, "hash1", "f1").Result()
				require.NoError(t, err)
				assert.Equal(t, "v1", v)
			},
		},
		{
			name: "HEXISTS",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash2", "f1", "v1").Err())
				ok, err := c.HExists(ctx, "hash2", "f1").Result()
				require.NoError(t, err)
				assert.True(t, ok)
			},
		},
		{
			name: "HINCRBY",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash3", "n", "1").Err())
				n, err := c.HIncrBy(ctx, "hash3", "n", 4).Result()
				require.NoError(t, err)
				assert.EqualValues(t, 5, n)
			},
		},
		{
			name: "HKEYS",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash4", "f1", "v1").Err())
				keys, err := c.HKeys(ctx, "hash4").Result()
				require.NoError(t, err)
				assert.Contains(t, keys, "f1")
			},
		},
		{
			name: "HVALS",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash5", "f1", "v1").Err())
				vals, err := c.HVals(ctx, "hash5").Result()
				require.NoError(t, err)
				assert.Contains(t, vals, "v1")
			},
		},
		{
			name: "HGETALL",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash6", "f1", "v1").Err())
				all, err := c.HGetAll(ctx, "hash6").Result()
				require.NoError(t, err)
				assert.Equal(t, "v1", all["f1"])
			},
		},
		{
			name: "HLEN",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash7", "f1", "v1").Err())
				n, err := c.HLen(ctx, "hash7").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 1, n)
			},
		},
		{
			name: "HDEL",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.HSet(ctx, "hash8", "f1", "v1").Err())
				n, err := c.HDel(ctx, "hash8", "f1").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 1, n)
			},
		},
	})
}

func TestCacheInstance_ListCommands(t *testing.T) {
	runInstanceCmdCases(t, []instanceCmdCase{
		{
			name: "LPUSH and LPOP",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.LPush(ctx, "list1", "a").Err())
				assertPrefixed(t, mr, "list1")
				v, err := c.LPop(ctx, "list1").Result()
				require.NoError(t, err)
				assert.Equal(t, "a", v)
			},
		},
		{
			name: "RPUSH, LLEN and RPOP",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.RPush(ctx, "list2", "x", "y").Err())
				n, err := c.LLen(ctx, "list2").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 2, n)
				v, err := c.RPop(ctx, "list2").Result()
				require.NoError(t, err)
				assert.Equal(t, "y", v)
			},
		},
	})
}

func TestCacheInstance_SetCommands(t *testing.T) {
	runInstanceCmdCases(t, []instanceCmdCase{
		{
			name: "SADD and SMEMBERS",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.SAdd(ctx, "set1", "m1", "m2").Err())
				assertPrefixed(t, mr, "set1")
				members, err := c.SMembers(ctx, "set1").Result()
				require.NoError(t, err)
				assert.ElementsMatch(t, []string{"m1", "m2"}, members)
			},
		},
		{
			name: "SCARD",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.SAdd(ctx, "set2", "m1", "m2").Err())
				n, err := c.SCard(ctx, "set2").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 2, n)
			},
		},
		{
			name: "SREM",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.SAdd(ctx, "set3", "m1").Err())
				n, err := c.SRem(ctx, "set3", "m1").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 1, n)
			},
		},
	})
}

func TestCacheInstance_SortedSetCommands(t *testing.T) {
	runInstanceCmdCases(t, []instanceCmdCase{
		{
			name: "ZADD and ZSCORE",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.ZAdd(ctx, "zset1", goredis.Z{Score: 1, Member: "m1"}).Err())
				assertPrefixed(t, mr, "zset1")
				score, err := c.ZScore(ctx, "zset1", "m1").Result()
				require.NoError(t, err)
				assert.Equal(t, float64(1), score)
			},
		},
		{
			name: "ZRANGE",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.ZAdd(ctx, "zset2", goredis.Z{Score: 1, Member: "m1"}, goredis.Z{Score: 2, Member: "m2"}).Err())
				members, err := c.ZRange(ctx, "zset2", 0, -1).Result()
				require.NoError(t, err)
				assert.Equal(t, []string{"m1", "m2"}, members)
			},
		},
		{
			name: "ZREM",
			run: func(t *testing.T, ctx context.Context, c *goredis.Client, mr *miniredis.Miniredis) {
				require.NoError(t, c.ZAdd(ctx, "zset3", goredis.Z{Score: 1, Member: "m1"}).Err())
				n, err := c.ZRem(ctx, "zset3", "m1").Result()
				require.NoError(t, err)
				assert.EqualValues(t, 1, n)
			},
		},
	})
}

func TestCacheInstance_Close(t *testing.T) {
	instance, _ := newTestInstance(t)
	assert.NoError(t, instance.Close())
}
