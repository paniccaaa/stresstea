package config

// TUIConfig holds TUI configuration
type TUIConfig struct {
	RefreshRate int    `yaml:"refresh_rate" default:"100"` // milliseconds
	Theme       string `yaml:"theme" default:"default"`    // default, dark, light
	ShowHelp    bool   `yaml:"show_help" default:"true"`
}

// DefaultTUIConfig returns default TUI configuration
func DefaultTUIConfig() *TUIConfig {
	return &TUIConfig{
		RefreshRate: 100,
		Theme:       "default",
		ShowHelp:    true,
	}
}
