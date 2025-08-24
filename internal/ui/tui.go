package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/paniccaaa/stresstea/internal/config"
	"github.com/paniccaaa/stresstea/internal/loadtest"
)

type TUI struct {
	config  *config.Config
	results []loadtest.Result
	start   time.Time
}

type model struct {
	config  *config.Config
	results []loadtest.Result
	start   time.Time
	width   int
	height  int
}

func NewTUI(cfg *config.Config) *TUI {
	return &TUI{
		config: cfg,
		start:  time.Now(),
	}
}

func (t *TUI) Run() error {
	m := model{
		config: t.config,
		start:  t.start,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func (t *TUI) UpdateResults(results chan loadtest.Result) {
	for result := range results {
		t.results = append(t.results, result)
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true).
		Render("Stresstea - Load Testing")

	configInfo := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		Render(fmt.Sprintf(`
Configuration:
  Target: %s
  Protocol: %s
  Duration: %v
  RPS: %d
  Threads: %d
`, m.config.Target, m.config.Protocol, m.config.Duration, m.config.Rate, m.config.Concurrent))

	stats := m.calculateStats()
	statsView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		Render(fmt.Sprintf(`
Statistics:
  Total Requests: %d
  Successful: %d
  Errors: %d
  Average Response Time: %v
  Max Response Time: %v
  Min Response Time: %v
`, stats.Total, stats.Success, stats.Errors, stats.AvgLatency, stats.MaxLatency, stats.MinLatency))

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("Press 'q' to exit")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		configInfo,
		"",
		statsView,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

type stats struct {
	Total      int
	Success    int
	Errors     int
	AvgLatency time.Duration
	MaxLatency time.Duration
	MinLatency time.Duration
}

func (m model) calculateStats() stats {
	if len(m.results) == 0 {
		return stats{}
	}

	s := stats{
		Total:      len(m.results),
		MinLatency: time.Hour, // Начальное значение
	}

	var totalLatency time.Duration
	for _, result := range m.results {
		if result.Error != nil {
			s.Errors++
		} else {
			s.Success++
		}

		if result.Latency < s.MinLatency {
			s.MinLatency = result.Latency
		}
		if result.Latency > s.MaxLatency {
			s.MaxLatency = result.Latency
		}

		totalLatency += result.Latency
	}

	if s.Success > 0 {
		s.AvgLatency = totalLatency / time.Duration(s.Success)
	}

	return s
}
