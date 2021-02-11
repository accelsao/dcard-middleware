package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"github.com/accelsao/dcard-middleware/middleware"
)

// NewServer create a new server
func NewServer() *http.Server {
	mux := http.NewServeMux()
	handler := middleware.NewMiddleware(1000)
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
