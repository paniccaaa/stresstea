package engine

import (
	"context"
	"fmt"

	"github.com/paniccaaa/stresstea/internal/config"
	"github.com/paniccaaa/stresstea/internal/loadtest"
	"github.com/paniccaaa/stresstea/internal/parser"
	"github.com/paniccaaa/stresstea/internal/ui"
	"go.uber.org/zap"
)

type Engine struct {
	config     *parser.Config
	logger     *zap.Logger
	outputMode string // "compact", "console"
}

func Run(cfg *parser.Config) error {
	return RunWithMode(cfg, "compact")
}

func RunWithMode(cfg *parser.Config, mode string) error {
	var logger *zap.Logger
	var err error

	// Use app logger config if available, otherwise use development logger
	if cfg.App != nil && cfg.App.Logger != nil {
		logger, err = config.NewLogger(cfg.App.Logger)
	} else {
		logger, err = config.NewDevelopmentLogger()
	}

	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	engine := &Engine{
		config:     cfg,
		logger:     logger,
		outputMode: mode,
	}

	ctx := context.Background()
	results := make(chan loadtest.Result, 1000)

	// Start load testing in background
	go func() {
		if err := engine.runLoadTest(ctx, results); err != nil {
			logger.Error("load test failed", zap.Error(err))
		}
	}()

	compactTUI := ui.NewCompactTUI(cfg)
	return compactTUI.Run(results)
}
func (e *Engine) runLoadTest(ctx context.Context, results chan<- loadtest.Result) error {
	var tester loadtest.LoadTester
	var err error

	switch e.config.Test.Protocol {
	case "http":
		tester, err = loadtest.NewHTTPTester(e.config)
	default:
		return fmt.Errorf("unsupported protocol: %s", e.config.Test.Protocol)
	}

	if err != nil {
		return fmt.Errorf("failed to create tester: %w", err)
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		if err := tester.Run(ctx, results); err != nil {
			e.logger.Error("test execution error", zap.Error(err))
		}
	}()

	<-done

	return nil
}
