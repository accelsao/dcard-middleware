package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"sort"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

var _ = fmt.Printf

var ctx = context.Background()

// Implements RedisClient for redis.Client
type redisClient struct {
	*redis.Client
}

func (c *redisClient) RateDel(key string) error {
	return c.Del(ctx, key).Err()
}

func (c *redisClient) RateEvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.EvalSha(ctx, sha1, keys, args...).Result()
}

func (c *redisClient) RateScriptLoad(script string) (string, error) {
	return c.ScriptLoad(ctx, script).Result()
}

type redisRingClient struct {
	*redis.Ring
}

func (r *redisRingClient) RateDel(key string) error {
	return r.Del(ctx, key).Err()
}

func (r *redisRingClient) RateEvalSha(sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return r.EvalSha(ctx, sha1, keys, args...).Result()
}

func (r *redisRingClient) RateScriptLoad(script string) (string, error) {
	var sha1 string
	err := r.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		res, err := shard.ScriptLoad(ctx, script).Result()
		if err == nil {
			sha1 = res
		}
		return err
	})
	return sha1, err
}

func TestRedisRateLimiter(t *testing.T) {

	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping(ctx).Result()
	if pong != "PONG" {
		t.Errorf("pong should be %v", "PONG")
	}
	if err != nil {
		t.Error(err)
	}

	defer rdb.Close()

	t.Run("ratelimiter.New, With Default Options", func(t *testing.T) {
		assert := assert.New(t)

		var limiter *Limiter
		var id = genID()
		t.Run("ratelimiter.New", func(t *testing.T) {
			limiter = New(Options{Client: &redisClient{rdb}})
		})
		t.Run("limiter.Get", func(t *testing.T) {
			res, err := limiter.Get(id)
			assert.Nil(err)
			assert.Equal(999, res.Remaining)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(998, res.Remaining)
			s.FastForward(time.Hour + time.Millisecond)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(999, res.Remaining)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(998, res.Remaining)
		})
		t.Run("limiter.Remove", func(t *testing.T) {
			res, err := limiter.Get(id)
			assert.Nil(err)
			assert.Equal(997, res.Remaining)
			err = limiter.Remove(id)
			assert.Nil(err)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(999, res.Remaining)
		})
	})

	t.Run("ratelimiter.New, With Options", func(t *testing.T) {
		assert := assert.New(t)

		var limiter *Limiter
		var id = genID()
		t.Run("ratelimiter.New", func(t *testing.T) {
			limiter = New(Options{
				IPLimit:  2,
				Duration: time.Second,
				Client:   &redisClient{rdb}})
		})

		t.Run("limiter.Get", func(t *testing.T) {
			res, err := limiter.Get(id)
			assert.Nil(err)
			assert.Equal(1, res.Remaining)
			s.FastForward(time.Millisecond * 500)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(0, res.Remaining)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(-1, res.Remaining)
			s.FastForward(time.Millisecond * 499)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(-1, res.Remaining)
			s.FastForward(time.Millisecond * 1)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(1, res.Remaining)
		})

		t.Run("limiter.Remove", func(t *testing.T) {
			res, err := limiter.Get(id)
			assert.Nil(err)
			assert.Equal(0, res.Remaining)
			err = limiter.Remove(id)
			assert.Nil(err)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(1, res.Remaining)
		})

	})

	t.Run("ratelimiter.New with redis ring", func(t *testing.T) {

		s1, err := miniredis.Run()
		if err != nil {
			panic(err)
		}
		defer s1.Close()
		s2, err := miniredis.Run()
		if err != nil {
			panic(err)
		}
		defer s2.Close()

		var limiter *Limiter
		t.Run("multi server for an IP address", func(t *testing.T) {
			assert := assert.New(t)
			var wg sync.WaitGroup
			var id = genID()
			var resp = NewRingResult(1000)
			var worker = func(r *redis.Ring, l *Limiter) {
				defer wg.Done()
				defer r.Close()
				for i := 0; i < 100; i++ {
					res, err := l.Get(id)
					assert.Nil(err)
					resp.Push(res.Remaining)
				}
			}
			wg.Add(10)
			for i := 0; i < 10; i++ {
				rdb := redis.NewRing(&redis.RingOptions{
					Addrs: map[string]string{
						"shard1": s1.Addr(),
						"shard2": s2.Addr(),
					},
				})
				limiter = New(Options{
					IPLimit:  997,
					Duration: time.Minute,
					Client:   &redisRingClient{rdb}})
				go worker(rdb, limiter)
			}
			wg.Wait()
			res := resp.Value()
			sort.Ints(res)
			assert.Equal(res[0], -1)
			assert.Equal(res[1], -1)
			for i := 2; i < 1000; i++ {
				assert.Equal(res[i], i-3)
			}

		})

	})

}

func genID() string {
	buf := make([]byte, 12)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

type RingResult struct {
	mu  sync.Mutex
	val []int
}

func NewRingResult(cap int) RingResult {
	return RingResult{
		val: make([]int, 0, cap),
	}
}

func (r *RingResult) Push(v int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.val = append(r.val, v)
}

func (r *RingResult) Value() []int {
	return r.val
}
