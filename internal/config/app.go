package config

// AppConfig holds the complete application configuration
type AppConfig struct {
	Logger *LoggerConfig `yaml:"logger"`
	TUI    *TUIConfig    `yaml:"tui"`
}

// DefaultAppConfig returns default application configuration
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Logger: &LoggerConfig{
			Level:      "info",
			Format:     "console",
			OutputPath: "stdout",
			ErrorPath:  "stderr",
		},
		TUI: DefaultTUIConfig(),
	}
}
