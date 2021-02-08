package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(t *testing.T) {
	handler := NewMiddleware()
	s := httptest.NewServer(handler)
	defer s.Close()

	for i := 0; i < 3; i++ {
		client := s.Client()
		fmt.Printf("server url: %s\n", s.URL)
		req, err := http.NewRequest("GET", s.URL, nil)
		req.Header.Add("X-Forwarded-For", "1.2.3.5:8080")

		resp, err := client.Do(req)
		// resp, err := http.Get(s.URL)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Receive %d\n", resp.StatusCode)
		}
		actual, err := ioutil.ReadAll(resp.Body)
		fmt.Printf("%v\n", string(actual))
	}
}
