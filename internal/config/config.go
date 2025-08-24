package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Target     string            `yaml:"target"`
	Duration   time.Duration     `yaml:"duration"`
	Rate       int               `yaml:"rate"`
	Concurrent int               `yaml:"concurrent"`
	Protocol   string            `yaml:"protocol"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	Body       string            `yaml:"body,omitempty"`
	Method     string            `yaml:"method,omitempty"`
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

	// Convert YAML configuration to Config
	config := &Config{
		Target:     yamlConfig.Global.Target,
		Duration:   yamlConfig.Global.Duration,
		Rate:       yamlConfig.Global.Rate,
		Concurrent: yamlConfig.Global.Concurrent,
		Protocol:   yamlConfig.Global.Protocol,
	}

	return config, nil
}
