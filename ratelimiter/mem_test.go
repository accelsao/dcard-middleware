package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemRateLimiter(t *testing.T) {
	t.Run("RateLimiter with default", func(t *testing.T) {
		assert := assert.New(t)

		limiter := New(Options{})
		id := genID()
		res, err := limiter.Get(id)
		assert.Nil(err)
		assert.Equal(res.Remaining, 999)
		assert.True(res.Reset.After(time.Now()))
		res, err = limiter.Get(id)
		assert.Nil(err)
		assert.Equal(res.Remaining, 998)
	})
	t.Run("RateLimiter expires", func(t *testing.T) {
		assert := assert.New(t)

		limiter := New(Options{
			IPLimit:  2,
			Duration: time.Second,
		})
		id := genID()
		res, err := limiter.Get(id)
		assert.Nil(err)
		assert.Equal(res.Remaining, 1)
		assert.True(res.Reset.After(time.Now()))

		// Wait until reset
		time.Sleep(time.Second)

		res, err = limiter.Get(id)
		assert.Nil(err)
		assert.Equal(res.Remaining, 1)
		assert.True(res.Reset.After(time.Now()))
	})
}
