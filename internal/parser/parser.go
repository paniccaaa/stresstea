package parser

import (
	"fmt"
	"os"
	"time"

	"github.com/paniccaaa/stresstea/internal/config"
	"gopkg.in/yaml.v3"
)

// TestRunConfig holds configuration for a single test run
type TestRunConfig struct {
	Target     string            `yaml:"target"`
	Duration   time.Duration     `yaml:"duration"`
	Rate       int               `yaml:"rate"`
	Concurrent int               `yaml:"concurrent"`
	Protocol   string            `yaml:"protocol"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	Body       string            `yaml:"body,omitempty"`
	Method     string            `yaml:"method,omitempty"`
}

// Config is the main configuration struct that combines all configs
type Config struct {
	App  *config.AppConfig `yaml:"app,omitempty"`
	Test *TestRunConfig    `yaml:"test"`
}

type YAMLConfig struct {
	Global    GlobalConfig     `yaml:"global"`
	Scenarios []ScenarioConfig `yaml:"scenarios"`
}

type GlobalConfig struct {
	Target     string        `yaml:"target"`
	Duration   time.Duration `yaml:"duration"`
	Rate       int           `yaml:"rate"`
	Concurrent int           `yaml:"concurrent"`
	Protocol   string        `yaml:"protocol"`
}

type ScenarioConfig struct {
	Name string       `yaml:"name"`
	Flow []StepConfig `yaml:"flow"`
}

type StepConfig struct {
	HTTP *HTTPStepConfig `yaml:"http,omitempty"`
	GRPC *GRPCStepConfig `yaml:"grpc,omitempty"`
	Wait *WaitStepConfig `yaml:"wait,omitempty"`
}

type HTTPStepConfig struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

type GRPCStepConfig struct {
	Service string                 `yaml:"service"`
	Method  string                 `yaml:"method"`
	Request map[string]interface{} `yaml:"request,omitempty"`
	Headers map[string]string      `yaml:"headers,omitempty"`
}

type WaitStepConfig struct {
	Duration time.Duration `yaml:"duration"`
}

func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var yamlConfig YAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Валидация конфигурации
	if err := validateYAMLConfig(&yamlConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Convert YAML configuration to Config
	config := &Config{
		App: config.DefaultAppConfig(),
		Test: &TestRunConfig{
			Target:     yamlConfig.Global.Target,
			Duration:   yamlConfig.Global.Duration,
			Rate:       yamlConfig.Global.Rate,
			Concurrent: yamlConfig.Global.Concurrent,
			Protocol:   yamlConfig.Global.Protocol,
		},
	}

	// TODO: Добавить поддержку scenarios в будущем
	// Пока используем только Global конфигурацию

	return config, nil
}

// validateYAMLConfig валидирует YAML конфигурацию
func validateYAMLConfig(config *YAMLConfig) error {
	if config.Global.Target == "" {
		return fmt.Errorf("target is required")
	}

	if config.Global.Rate <= 0 {
		return fmt.Errorf("rate must be positive")
	}

	if config.Global.Concurrent <= 0 {
		return fmt.Errorf("concurrent must be positive")
	}

	if config.Global.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	if config.Global.Protocol != "http" && config.Global.Protocol != "grpc" {
		return fmt.Errorf("protocol must be 'http' or 'grpc'")
	}

	return nil
}
