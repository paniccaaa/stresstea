package loadtest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/paniccaaa/stresstea/internal/parser"
)

type HTTPTester struct {
	*BaseTester
	client *http.Client
}

func NewHTTPTester(cfg *parser.Config) (*HTTPTester, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Test.Concurrent,
			MaxIdleConnsPerHost: cfg.Test.Concurrent,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HTTPTester{
		BaseTester: NewBaseTester(cfg),
		client:     client,
	}, nil
}

func (h *HTTPTester) Run(ctx context.Context, results chan<- Result) error {
	defer close(results)

	var wg sync.WaitGroup
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < h.config.Test.Concurrent; i++ {
		wg.Add(1)
		go h.worker(workerCtx, &wg, results)
	}

	timer := time.NewTimer(h.config.Test.Duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		cancel()
		wg.Wait()
		return nil
	case <-ctx.Done():
		cancel()
		wg.Wait()
		return ctx.Err()
	}
}

func (h *HTTPTester) worker(ctx context.Context, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()

	ratePerWorker := h.config.Test.Rate / h.config.Test.Concurrent
	if ratePerWorker <= 0 {
		ratePerWorker = 1
	}

	ticker := time.NewTicker(time.Second / time.Duration(ratePerWorker))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result := h.makeRequest()
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (h *HTTPTester) makeRequest() Result {
	start := time.Now()

	method := h.config.Test.Method
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if h.config.Test.Body != "" {
		body = strings.NewReader(h.config.Test.Body)
	}

	req, err := http.NewRequest(method, h.config.Test.Target, body)
	if err != nil {
		return Result{
			Timestamp: start,
			Latency:   time.Since(start),
			Error:     fmt.Errorf("failed to create request: %w", err),
		}
	}

	for k, v := range h.config.Test.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return Result{
			Timestamp: start,
			Latency:   time.Since(start),
			Error:     fmt.Errorf("failed to execute request: %w", err),
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{
			Timestamp: start,
			Latency:   time.Since(start),
			Error:     fmt.Errorf("failed to read response: %w", err),
		}
	}

	return Result{
		Timestamp: start,
		Latency:   time.Since(start),
		Status:    resp.StatusCode,
		Bytes:     int64(len(bodyBytes)),
	}
}
