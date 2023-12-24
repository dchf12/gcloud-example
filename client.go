package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

type loggingRoundTripper struct {
	transport http.RoundTripper
	logger    func(string, ...any)
}

func (t *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.logger == nil {
		t.logger = log.Printf
	}
	start := time.Now()
	resp, err := t.transport.RoundTrip(req)
	if resp != nil {
		t.logger("%s %s %d %s, duration: %d", req.Method, req.URL, resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start))
	}
	return resp, err
}

type basicAuthRoundTripper struct {
	username string
	password string
	base     http.RoundTripper
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.username, rt.password)
	return rt.base.RoundTrip(req)
}

type retryableRoundTripper struct {
	base     http.RoundTripper
	attempts int
	waitTime time.Duration
}

func (rt *retryableRoundTripper) shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}
	}

	if resp != nil {
		if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode <= 504) {
			return true
		}
	}
	return false
}

func (rt *retryableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for count := 0; count < rt.attempts; count++ {
		if !rt.shouldRetry(resp, err) {
			return resp, err
		}
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(rt.waitTime):
		}
	}
	return resp, err
}

func main() {
	client := &http.Client{
		Transport: &loggingRoundTripper{
			transport: http.DefaultTransport,
		},
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	client2 := &http.Client{
		Transport: &basicAuthRoundTripper{
			username: "user",
			password: "pass",
			base:     http.DefaultTransport,
		},
	}
	ctx2 := context.Background()
	req2, err := http.NewRequestWithContext(ctx2, http.MethodGet, "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp2, err := client2.Do(req2)
	if err != nil {
		log.Fatal(err)
	}
	defer resp2.Body.Close()
	client3 := &http.Client{
		Transport: &retryableRoundTripper{
			base:     http.DefaultTransport,
			attempts: 3,
			waitTime: 1 * time.Second,
		},
	}
	ctx3 := context.Background()
	req3, err := http.NewRequestWithContext(ctx3, http.MethodGet, "http://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp3, err := client3.Do(req3)
	if err != nil {
		log.Fatal(err)
	}
	defer resp3.Body.Close()
}
