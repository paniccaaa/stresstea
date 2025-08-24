package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `yaml:"level" default:"info"`
	Format     string `yaml:"format" default:"json"` // json or console
	OutputPath string `yaml:"output_path" default:"stdout"`
	ErrorPath  string `yaml:"error_path" default:"stderr"`
}

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{cfg.OutputPath}
	config.ErrorOutputPaths = []string{cfg.ErrorPath}

	if cfg.Format == "console" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	return config.Build()
}

// NewDevelopmentLogger creates a development logger suitable for TUI
func NewDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}

	return config.Build()
}
