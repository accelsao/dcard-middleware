package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// ClientMetadata is recorded for server usages
type ClientMetadata struct {
	mu         sync.Mutex
	visit      map[string]int
	remainTime int // unit: second
}

// NewMiddleware creates a handler
func NewMiddleware() *ClientMetadata {
	mw := &ClientMetadata{
		visit:      make(map[string]int),
		remainTime: 3600,
	}
	return mw
}

func (h *ClientMetadata) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addr := GetIP(r)
	var num int
	h.mu.Lock()
	if _, exist := h.visit[addr]; !exist {
		h.visit[addr] = 0
	}
	h.visit[addr]++
	num = h.visit[addr]
	h.mu.Unlock()

	fmt.Fprintf(w, "client from %s visit %d times\n", addr, num)
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
	// mux.ListenAndServe(":8080", nil)
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return s
}

func main() {
	s := NewServer()
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("server failed to start with error %v", err.Error())
	}
}
