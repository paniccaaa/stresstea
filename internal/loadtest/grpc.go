package loadtest

import (
	"context"
	"fmt"
	"time"

	"github.com/paniccaaa/stresstea/internal/parser"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCTester struct {
	*BaseTester
	conn *grpc.ClientConn
}

func NewGRPCTester(cfg *parser.Config) (*GRPCTester, error) {
	conn, err := grpc.NewClient(cfg.Test.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GRPCTester{
		BaseTester: NewBaseTester(cfg),
		conn:       conn,
	}, nil
}

func (g *GRPCTester) Run(ctx context.Context, results chan<- Result) error {
	defer close(results)
	defer g.conn.Close()

	// TODO: Implement gRPC load testing
	// This requires dynamic method invocation through reflection
	// or pre-generated code from .proto files

	// Temporary placeholder
	time.Sleep(g.config.Test.Duration)

	return fmt.Errorf("gRPC testing not implemented yet")
}
