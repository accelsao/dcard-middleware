package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	RateLimit   = 2
	TimeToReset = time.Duration(1) * time.Hour
)

type Metadata struct {
	count      int
	timeExpire time.Time
}

// ClientMetadata is recorded for server usages
type ClientMetadata struct {
	mu    sync.Mutex
	visit map[string]Metadata
}

// NewMiddleware creates a handler
func NewMiddleware() *ClientMetadata {
	mw := &ClientMetadata{
		visit: make(map[string]Metadata),
	}
	return mw
}

func (h *ClientMetadata) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := time.Now()

	addr := GetIP(r)
	var data Metadata
	h.mu.Lock()
	if data, exist := h.visit[addr]; !exist {
		h.visit[addr] = Metadata{
			count:      1,
			timeExpire: GetNextExpireTime(t),
		}
	} else {
		// check whether visit times reach limit
		if data.count >= RateLimit {
			if t.Sub(data.timeExpire) >= 0 {
				h.visit[addr] = Metadata{
					count:      1,
					timeExpire: GetNextExpireTime(t),
				}
			} else {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
		} else {
			// TODO: golang cannot assign to struct field in map, is there idiomatic way to implement?
			h.visit[addr] = Metadata{
				count:      data.count + 1,
				timeExpire: data.timeExpire,
			}
		}
	}
	data = h.visit[addr]

	h.mu.Unlock()

	fmt.Fprintf(w, "client from %s visit %d times, reset after %v\n", addr, data.count, data.timeExpire)
}

// GetNextExpireTime return time round to next hour
func GetNextExpireTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
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
	handler := NewMiddleware()
	mux.Handle("/hello", handler)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return s
}

func main() {
	s := NewServer()

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := s.Shutdown(context.Background()); err != nil {
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
