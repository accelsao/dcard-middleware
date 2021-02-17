package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/accelsao/dcard-middleware/middleware"
	"github.com/go-redis/redis/v8"
)

var _ = fmt.Printf
var ctx = context.Background()
var use_mem bool

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

// GetIP get client IP address
// 1. Check Header "X-Forwarded-For"
// 2. Check RemoteAddress
func GetIP(r *http.Request) string {
	addr := r.Header.Get("X-Forwarded-For")
	if addr != "" {
		return addr
	}
	return r.RemoteAddr
}

// NewServer create a new server
func NewServer() *http.Server {
	mux := http.NewServeMux()

	var limiter *middleware.Limiter
	if use_mem {
		limiter = middleware.New(middleware.Options{
			IPLimit:  2,
			Duration: time.Second * 10,
		})

	} else {
		client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		limiter = middleware.New(middleware.Options{
			IPLimit:  2,
			Duration: time.Minute,
			Client:   &redisClient{client},
		})
	}

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		ip := GetIP(r)
		res, err := limiter.Get(ip)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		header := w.Header()
		header.Set("X-Ratelimit-Remaining", strconv.FormatInt(int64(res.Remaining), 10))
		header.Set("X-Ratelimit-Reset", res.Reset.String())

		if res.Remaining >= 0 {
			w.WriteHeader(200)
		} else {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		}
	})

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return s
}

func main() {

	useMemFlag := flag.Bool("use_mem", false, "set to use memory")
	flag.Parse()
	use_mem = *useMemFlag

	s := NewServer()

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := s.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
