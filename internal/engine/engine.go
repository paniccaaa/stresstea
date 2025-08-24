package engine

import (
	"context"
	"fmt"

	"github.com/paniccaaa/stresstea/internal/config"
	"github.com/paniccaaa/stresstea/internal/loadtest"
	"github.com/paniccaaa/stresstea/internal/ui"
	"go.uber.org/zap"
)

type Engine struct {
	config *config.Config
	logger *zap.Logger
	ui     *ui.TUI
}

func Run(cfg *config.Config) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	engine := &Engine{
		config: cfg,
		logger: logger,
	}

	engine.ui = ui.NewTUI(cfg)

	ctx := context.Background()
	go func() {
		if err := engine.runLoadTest(ctx); err != nil {
			logger.Error("load test failed", zap.Error(err))
		}
	}()

	if err := engine.ui.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}

func (e *Engine) runLoadTest(ctx context.Context) error {
	var tester loadtest.LoadTester
	var err error

	switch e.config.Protocol {
	case "http":
		tester, err = loadtest.NewHTTPTester(e.config)
	case "grpc":
		tester, err = loadtest.NewGRPCTester(e.config)
	default:
		return fmt.Errorf("unsupported protocol: %s", e.config.Protocol)
	}

	if err != nil {
		return fmt.Errorf("failed to create tester: %w", err)
	}

	results := make(chan loadtest.Result, 1000)
	done := make(chan bool)

	go func() {
		defer close(done)
		if err := tester.Run(ctx, results); err != nil {
			e.logger.Error("test execution error", zap.Error(err))
		}
	}()

	go e.ui.UpdateResults(results)

	<-done

	return nil
}
