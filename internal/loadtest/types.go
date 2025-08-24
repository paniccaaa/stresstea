package loadtest

import (
	"context"
	"time"

	"github.com/paniccaaa/stresstea/internal/config"
)

type Result struct {
	Timestamp time.Time
	Latency   time.Duration
	Error     error
	Status    int
	Bytes     int64
}

type LoadTester interface {
	Run(ctx context.Context, results chan<- Result) error
}

type BaseTester struct {
	config *config.Config
}

func NewBaseTester(cfg *config.Config) *BaseTester {
	return &BaseTester{
		config: cfg,
	}
}
