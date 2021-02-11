package main

import (
	"fmt"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

const (
	// RateLimit   = 10
	TimeToReset = time.Duration(1) * time.Hour
)

type Metadata struct {
	count      int
	timeExpire time.Time
}

// ServerHandler is recorded for server usages
type ServerHandler struct {
	mu        sync.Mutex
	visit     map[string]Metadata
	ratelimit int
}

// NewMiddleware creates a handler
func NewMiddleware(ratelimit int) *ServerHandler {
	mw := &ServerHandler{
		visit:     make(map[string]Metadata),
		ratelimit: ratelimit,
	}
	return mw
}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	t := GetCurrentTime(r)
	// fmt.Printf("time:%v\n", t)	
	addr := GetIP(r)
	// fmt.Printf("addr:%v\n", addr)	
	var data Metadata
	h.mu.Lock()
	if data, exist := h.visit[addr]; !exist {
		fmt.Println("Here")
		h.visit[addr] = Metadata{
			count:      1,
			timeExpire: GetNextExpireTime(t),
		}
	} else {
		fmt.Printf("data.count: %v, h.ratelimiet: %v\n", data.count, h.ratelimit)

		// check whether visit times reach limit
		if data.count >= h.ratelimit {
			// timeElapsed := t.Sub(data.timeExpire)
			fmt.Printf("t: %v -> data.timeExpire: %v\n", t, data.timeExpire)
			after := t.After(data.timeExpire) || t.Equal(data.timeExpire)
			fmt.Printf("after: %v\n", after)
			if after {
				h.visit[addr] = Metadata{
					count:      1,
					timeExpire: GetNextExpireTime(t),
				}
			} else {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				h.mu.Unlock()
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

	// Add header
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(h.ratelimit-data.count))
	w.Header().Set("X-RateLimit-Reset", data.timeExpire.Sub(t).String())

	// fmt.Fprintf(w, "[log] client from %s visit %d times, reset after %v\n", addr, data.count, data.timeExpire)
}

// GetCurrentTime return current time, Header.Date if in test mode
func GetCurrentTime(r *http.Request) time.Time {
	t := r.Header.Get("Date")
	if t != "" {
		// format: 2021-01-01 12:00:00 +0800 CST
		tm, _ := time.Parse(time.RFC3339, t)
		return tm
	}
	return time.Now()
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
	handler := NewMiddleware(1000)
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
