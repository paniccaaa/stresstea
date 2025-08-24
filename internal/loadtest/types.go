package loadtest

import (
	"context"
	"time"

	"github.com/paniccaaa/stresstea/internal/parser"
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
	config *parser.Config
}

func NewBaseTester(cfg *parser.Config) *BaseTester {
	return &BaseTester{
		config: cfg,
	}
}
