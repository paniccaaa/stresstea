package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/paniccaaa/stresstea/internal/loadtest"
	"github.com/paniccaaa/stresstea/internal/parser"
)

// TestStatus представляет статус теста
type TestStatus int

const (
	StatusRunning TestStatus = iota
	StatusPaused
	StatusStopped
)

// Константы для метрик
const (
	MaxResults    = 1000
	MaxRPSHistory = 60
	MaxErrors     = 10
)

// CompactTUI представляет компактный TUI интерфейс
type CompactTUI struct {
	config      *parser.Config
	metrics     *Metrics
	results     []loadtest.Result
	start       time.Time
	width       int
	height      int
	resultsChan chan loadtest.Result
	status      TestStatus
	showHelp    bool
}

// NewCompactTUI создает новый компактный TUI
func NewCompactTUI(cfg *parser.Config) *CompactTUI {
	return &CompactTUI{
		config:  cfg,
		metrics: NewMetrics(cfg),
		start:   time.Now(),
		status:  StatusRunning,
	}
}

// Run запускает компактный TUI
func (t *CompactTUI) Run(resultsChan chan loadtest.Result) error {
	t.resultsChan = resultsChan

	p := tea.NewProgram(t, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// Init инициализирует модель
func (t CompactTUI) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		t.waitForResults(),
	)
}

// Update обновляет состояние
func (t CompactTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		case "h":
			t.showHelp = !t.showHelp
		case "p":
			switch t.status {
			case StatusRunning:
				t.status = StatusPaused
			case StatusPaused:
				t.status = StatusRunning
			}
		}
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
	case loadtest.Result:
		if t.status == StatusRunning {
			t.results = append(t.results, msg)
			if len(t.results) > 1000 {
				t.results = t.results[len(t.results)-1000:]
			}
			t.metrics.UpdateMetrics(t.results)
		}
		return t, t.waitForResults()
	case time.Time:
		return t, t.waitForResults()
	}

	return t, nil
}

// View отображает компактный интерфейс
func (t CompactTUI) View() string {
	if t.width == 0 {
		return "Initializing..."
	}

	if t.showHelp {
		return t.renderHelp()
	}

	return t.renderCompactView()
}

// renderCompactView отображает компактный вид
func (t CompactTUI) renderCompactView() string {
	// Заголовок
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Margin(0, 0, 1, 0).
		Render("🚀 Stresstea - HTTP Load Testing")

	// Основная информация в одну строку
	mainInfo := t.renderMainInfo()

	// Метрики в компактном виде
	metrics := t.renderCompactMetrics()

	// Прогресс
	progress := t.renderProgress()

	// Статус коды (если есть данные)
	statusCodes := t.renderStatusCodes()

	// Ошибки (если есть)
	errors := t.renderErrors()

	// Помощь
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Press 'h' for help, 'q' to quit, 'p' to pause")

	// Собираем все вместе
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		mainInfo,
		metrics,
		progress,
		statusCodes,
		errors,
		"",
		help,
	)

	return content
}

// renderMainInfo отображает основную информацию
func (t CompactTUI) renderMainInfo() string {

	var statusColor, statusText string
	switch t.status {
	case StatusRunning:
		statusColor, statusText = "#00FF00", "RUNNING"
	case StatusPaused:
		statusColor, statusText = "#FFA500", "PAUSED"
	case StatusStopped:
		statusColor, statusText = "#FF0000", "STOPPED"
	}

	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true).
		Render(fmt.Sprintf("[%s]", statusText))

	// Целевой URL
	target := lipgloss.NewStyle().
		Bold(true).
		Render(fmt.Sprintf("Target: %s", t.config.Test.Target))

	// Параметры теста
	params := fmt.Sprintf("Rate: %d RPS | Threads: %d | Duration: %v",
		t.config.Test.Rate,
		t.config.Test.Concurrent,
		t.config.Test.Duration)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		status,
		" ",
		target,
		" | ",
		params,
	)
}

// renderCompactMetrics отображает метрики в компактном виде
func (t CompactTUI) renderCompactMetrics() string {
	// RPS и Success Rate в одну строку
	rps := fmt.Sprintf("RPS: %.1f/%d", t.metrics.CurrentRPS, t.metrics.TargetRPS)
	success := fmt.Sprintf("Success: %.1f%%", t.metrics.SuccessRate)

	// Latency метрики
	latency := fmt.Sprintf("Avg: %s | P90: %s | P99: %s",
		t.formatDuration(t.metrics.AvgLatency),
		t.formatDuration(t.metrics.P90Latency),
		t.formatDuration(t.metrics.P99Latency))

	// Requests и Errors
	requests := fmt.Sprintf("Requests: %d | Errors: %d",
		t.metrics.TotalRequests,
		t.metrics.FailedRequests)

	// Throughput
	throughput := fmt.Sprintf("Throughput: %.2f MB/s",
		t.metrics.ThroughputMBps)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, rps, " | ", success),
		latency,
		requests,
		throughput,
	)
}

// renderProgress отображает прогресс
func (t CompactTUI) renderProgress() string {
	progress := t.metrics.GetProgress()
	elapsed := t.metrics.ElapsedTime.Round(time.Second)
	remaining := t.metrics.RemainingTime.Round(time.Second)

	// Создаем простой прогресс бар
	barWidth := 40
	filled := int(float64(barWidth) * progress)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	progressBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Render(fmt.Sprintf("[%s] %.1f%%", bar, progress*100))

	timeInfo := fmt.Sprintf("Elapsed: %v | Remaining: %v", elapsed, remaining)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		progressBar,
		timeInfo,
	)
}

// renderStatusCodes отображает статус коды
func (t CompactTUI) renderStatusCodes() string {
	if len(t.metrics.StatusCodes) == 0 {
		return ""
	}

	// Используем отсортированные статус коды для стабильного отображения
	sortedCodes := t.metrics.GetStatusCodesSorted()

	var codes []string
	for _, codeInfo := range sortedCodes {
		style := GetStatusStyle(codeInfo.Status)
		codes = append(codes, style.Render(fmt.Sprintf("%d: %d", codeInfo.Status, codeInfo.Count)))
	}

	if len(codes) > 0 {
		return lipgloss.NewStyle().
			Bold(true).
			Render("Status Codes: ") + strings.Join(codes, " | ")
	}

	return ""
}

// renderErrors отображает ошибки
func (t CompactTUI) renderErrors() string {
	if len(t.metrics.RecentErrors) == 0 {
		return ""
	}

	// Показываем только последние 3 ошибки в стабильном порядке
	errors := make([]string, len(t.metrics.RecentErrors))
	copy(errors, t.metrics.RecentErrors)

	if len(errors) > 3 {
		errors = errors[len(errors)-3:]
	}

	errorText := "Errors: " + strings.Join(errors, " | ")
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Render(errorText)
}

// renderHelp отображает справку
func (t CompactTUI) renderHelp() string {
	helpText := `Stresstea - HTTP Load Testing Tool

Controls:
  h - Toggle help
  p - Pause/Resume test
  q - Quit application
  Ctrl+C - Force quit

Metrics:
  RPS - Requests per second
  Success - Success rate percentage
  Avg/P90/P99 - Response time percentiles
  Throughput - Data transfer rate

Press 'h' to close help`

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Render(helpText)
}

// waitForResults ожидает новые результаты
func (t CompactTUI) waitForResults() tea.Cmd {
	return func() tea.Msg {
		if t.resultsChan != nil {
			select {
			case result := <-t.resultsChan:
				return result
			default:
				return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
					return t
				})()
			}
		}
		return nil
	}
}

// formatDuration форматирует длительность
func (t CompactTUI) formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1e6)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
