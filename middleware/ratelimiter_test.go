package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

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
			assert.True(res.Reset.UnixNano() > time.Now().UnixNano())
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(998, res.Remaining)
		})
		t.Run("limiter.Remove", func(t *testing.T) {
			err := limiter.Remove(id)
			assert.Nil(err)
			err = limiter.Remove(id)
			assert.Nil(err)
			res, err := limiter.Get(id)
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
			assert.True(res.Reset.UnixNano() > time.Now().UnixNano())
			assert.True(res.Reset.UnixNano() <= time.Now().Add(time.Second).UnixNano())

			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(0, res.Remaining)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(-1, res.Remaining)
			res, err = limiter.Get(id)
			assert.Nil(err)
			assert.Equal(-1, res.Remaining)
		})
		t.Run("limiter.Remove", func(t *testing.T) {
			err := limiter.Remove(id)
			assert.Nil(err)
			err = limiter.Remove(id)
			assert.Nil(err)
			res, err := limiter.Get(id)
			assert.Nil(err)
			assert.Equal(1, res.Remaining)
		})
	})

}

func genID() string {
	buf := make([]byte, 12)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}
