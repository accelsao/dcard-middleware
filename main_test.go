package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
	"github.com/accelsao/dcard-middleware/middleware"
)

func TestSeqentialRequest(t *testing.T) {
	handler := middleware.NewMiddleware(10)
	
	s := httptest.NewServer(handler)
	defer s.Close()
	okCount := 0
	for i := 0; i <= handler.Ratelimit; i++ {
		client := s.Client()
		req, err := http.NewRequest("GET", s.URL, nil)
		// send request the same time from the same IP address
		req.Header.Add("X-Forwarded-For", "1.2.3.4:8080")
		req.Header.Add("Date", time.Date(2021, 1, 1, 12, 0, 0, 0, time.Local).Format(time.RFC3339))
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if i < handler.Ratelimit {
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Receive %d\n", resp.StatusCode)
			}
			ratelimitRemain, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
			if err != nil {
				t.Fatal(err)
			}
			if ratelimitRemain != handler.Ratelimit-(i+1) {
				t.Fatalf("X-RateLimit-Remaining should be %v, got %v\n", handler.Ratelimit-(i+1), ratelimitRemain)

			}
			okCount++
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("Receive %d\n", resp.StatusCode)
			}
		}

	}
	if okCount != handler.Ratelimit {
		t.Fatalf("number of success request should be %v, got %v\n", handler.Ratelimit, okCount)
	}
}

// Half of them send reqs on @time
// Other send reqs on @time + 1hr
func TestSeqentialRequest2(t *testing.T) {
	handler := middleware.NewMiddleware(20)
	s := httptest.NewServer(handler)
	defer s.Close()
	okCount := 0
	// 3 additional that must be blocked
	for i := 0; i < handler.Ratelimit+3; i++ {
		client := s.Client()
		req, err := http.NewRequest("GET", s.URL, nil)
		// send request the same time from the same IP address
		req.Header.Add("X-Forwarded-For", "1.2.3.4:8080")
		req.Header.Add("Date", time.Date(2021, 1, 1, 12, 0, 0, 0, time.Local).Format(time.RFC3339))
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if i < handler.Ratelimit {
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Receive %d\n", resp.StatusCode)
			}
			ratelimitRemain, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
			if err != nil {
				t.Fatal(err)
			}
			if ratelimitRemain != handler.Ratelimit-(i+1) {
				t.Fatalf("X-RateLimit-Remaining should be %v, got %v\n", handler.Ratelimit-(i+1), ratelimitRemain)

			}
			okCount++
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("Receive %d\n", resp.StatusCode)
			}
		}
	}
	for i := 0; i < handler.Ratelimit+3; i++ {
		client := s.Client()
		req, err := http.NewRequest("GET", s.URL, nil)
		// send request the same time from the same IP address
		req.Header.Add("X-Forwarded-For", "1.2.3.4:8080")
		req.Header.Add("Date", time.Date(2021, 1, 1, 13, 0, 0, 0, time.Local).Format(time.RFC3339))
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if i < handler.Ratelimit {
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Receive %d\n", resp.StatusCode)
			}
			ratelimitRemain, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
			if err != nil {
				t.Fatal(err)
			}
			if ratelimitRemain != handler.Ratelimit-(i+1) {
				t.Fatalf("X-RateLimit-Remaining should be %v, got %v\n", handler.Ratelimit-(i+1), ratelimitRemain)

			}
			okCount++
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("Receive %d\n", resp.StatusCode)
			}
		}
	}
	expectCount := handler.Ratelimit * 2
	if okCount != expectCount {
		t.Fatalf("number of success request should be %v, got %v\n", expectCount, okCount)
	}
}
