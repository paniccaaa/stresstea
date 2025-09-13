package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MetricsPanel отображает статистику в реальном времени с интерактивными элементами
type MetricsPanel struct {
	BaseComponent
	metrics *Metrics

	// Компоненты
	rpsProgress     progress.Model
	successProgress progress.Model
	metricsTable    table.Model

	// Состояние
	showDetails bool
}

// NewMetricsPanel создает новую панель метрик
func NewMetricsPanel(metrics *Metrics) *MetricsPanel {
	mp := &MetricsPanel{
		BaseComponent: BaseComponent{},
		metrics:       metrics,
		showDetails:   false,
	}

	mp.initComponents()
	return mp
}

// initComponents инициализирует компоненты
func (mp *MetricsPanel) initComponents() {
	// RPS Progress
	mp.rpsProgress = progress.New(progress.WithDefaultGradient())
	mp.rpsProgress.Width = 20
	mp.rpsProgress.ShowPercentage = false

	// Success Progress
	mp.successProgress = progress.New(progress.WithDefaultGradient())
	mp.successProgress.Width = 20
	mp.successProgress.ShowPercentage = true

	// Metrics Table
	mp.initMetricsTable()
}

// initMetricsTable инициализирует таблицу метрик
func (mp *MetricsPanel) initMetricsTable() {
	columns := []table.Column{
		{Title: "Metric", Width: 12},
		{Title: "Value", Width: 15},
		{Title: "Status", Width: 8},
	}

	rows := []table.Row{
		{"RPS", "0.0", "●"},
		{"Success", "0.0%", "●"},
		{"Avg Latency", "0ms", "●"},
		{"P90 Latency", "0ms", "●"},
		{"P95 Latency", "0ms", "●"},
		{"P99 Latency", "0ms", "●"},
		{"Errors", "0", "●"},
		{"Throughput", "0 MB/s", "●"},
	}

	mp.metricsTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(8),
	)

	// Стилизация таблицы
	mp.metricsTable.SetStyles(table.Styles{
		Header: TableHeaderStyle,
		Cell:   TableRowStyle,
	})
}

// Render отображает панель метрик
func (mp *MetricsPanel) Render(width, height int) string {
	mp.SetSize(width, height)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Render("Real-time Metrics")

	// Обновляем прогресс бары
	mp.updateProgressBars()

	// RPS секция
	rpsLabel := lipgloss.NewStyle().Bold(true).Render("Requests Per Second:")
	rpsValue := fmt.Sprintf("%.1f / %d", mp.metrics.CurrentRPS, mp.metrics.TargetRPS)
	rpsBar := mp.rpsProgress.View()
	rpsStatus := mp.getRPSStatus()

	// Success секция
	successLabel := lipgloss.NewStyle().Bold(true).Render("Success Rate:")
	successValue := fmt.Sprintf("%.1f%%", mp.metrics.SuccessRate)
	successBar := mp.successProgress.View()
	successStatus := mp.getSuccessStatus()

	// Основные метрики
	metricsContent := mp.renderMetricsTable()

	// Дополнительная информация
	additionalInfo := mp.renderAdditionalInfo()

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Press 'd' for details | 'r' to refresh")

	content := fmt.Sprintf(`%s

%s
%s %s %s

%s
%s %s %s

%s

%s

%s`,
		title,
		rpsLabel, rpsValue, rpsBar, rpsStatus,
		successLabel, successValue, successBar, successStatus,
		metricsContent,
		additionalInfo,
		helpText)

	return PanelStyle.Render(content)
}

// updateProgressBars обновляет прогресс бары
func (mp *MetricsPanel) updateProgressBars() {
	// RPS прогресс (относительно целевого RPS)
	rpsPercent := 0.0
	if mp.metrics.TargetRPS > 0 {
		rpsPercent = mp.metrics.CurrentRPS / float64(mp.metrics.TargetRPS)
		if rpsPercent > 1.0 {
			rpsPercent = 1.0
		}
	}
	mp.rpsProgress.SetPercent(rpsPercent)

	// Success прогресс
	successPercent := mp.metrics.SuccessRate / 100.0
	mp.successProgress.SetPercent(successPercent)
}

// getRPSStatus возвращает статус RPS
func (mp *MetricsPanel) getRPSStatus() string {
	if mp.metrics.TargetRPS == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("●")
	}

	ratio := mp.metrics.CurrentRPS / float64(mp.metrics.TargetRPS)

	switch {
	case ratio >= 0.9:
		return SuccessStyle.Render("●")
	case ratio >= 0.7:
		return WarningStyle.Render("●")
	default:
		return ErrorStyle.Render("●")
	}
}

// getSuccessStatus возвращает статус успешности
func (mp *MetricsPanel) getSuccessStatus() string {
	switch {
	case mp.metrics.SuccessRate >= 95:
		return SuccessStyle.Render("●")
	case mp.metrics.SuccessRate >= 80:
		return WarningStyle.Render("●")
	default:
		return ErrorStyle.Render("●")
	}
}

// renderMetricsTable отображает таблицу метрик
func (mp *MetricsPanel) renderMetricsTable() string {
	// Обновляем данные в таблице
	mp.updateTableData()

	tableContent := mp.metricsTable.View()

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Render(tableContent)
}

// updateTableData обновляет данные в таблице
func (mp *MetricsPanel) updateTableData() {
	rows := []table.Row{
		{"RPS", fmt.Sprintf("%.1f", mp.metrics.CurrentRPS), mp.getRPSStatus()},
		{"Success", fmt.Sprintf("%.1f%%", mp.metrics.SuccessRate), mp.getSuccessStatus()},
		{"Avg Latency", mp.formatDuration(mp.metrics.AvgLatency), mp.getLatencyStatus(mp.metrics.AvgLatency)},
		{"P90 Latency", mp.formatDuration(mp.metrics.P90Latency), mp.getLatencyStatus(mp.metrics.P90Latency)},
		{"P95 Latency", mp.formatDuration(mp.metrics.P95Latency), mp.getLatencyStatus(mp.metrics.P95Latency)},
		{"P99 Latency", mp.formatDuration(mp.metrics.P99Latency), mp.getLatencyStatus(mp.metrics.P99Latency)},
		{"Errors", strconv.Itoa(mp.metrics.FailedRequests), mp.getErrorStatus()},
		{"Throughput", fmt.Sprintf("%.2f MB/s", mp.metrics.ThroughputMBps), mp.getThroughputStatus()},
	}

	mp.metricsTable.SetRows(rows)
}

// getLatencyStatus возвращает статус задержки
func (mp *MetricsPanel) getLatencyStatus(latency time.Duration) string {
	switch {
	case latency < 100*time.Millisecond:
		return SuccessStyle.Render("●")
	case latency < 500*time.Millisecond:
		return WarningStyle.Render("●")
	default:
		return ErrorStyle.Render("●")
	}
}

// getErrorStatus возвращает статус ошибок
func (mp *MetricsPanel) getErrorStatus() string {
	errorRate := mp.metrics.ErrorRate
	switch {
	case errorRate < 1:
		return SuccessStyle.Render("●")
	case errorRate < 5:
		return WarningStyle.Render("●")
	default:
		return ErrorStyle.Render("●")
	}
}

// getThroughputStatus возвращает статус пропускной способности
func (mp *MetricsPanel) getThroughputStatus() string {
	// Простая логика на основе количества байт в секунду
	switch {
	case mp.metrics.BytesPerSecond > 1024*1024: // > 1MB/s
		return SuccessStyle.Render("●")
	case mp.metrics.BytesPerSecond > 100*1024: // > 100KB/s
		return WarningStyle.Render("●")
	default:
		return ErrorStyle.Render("●")
	}
}

// formatDuration форматирует длительность
func (mp *MetricsPanel) formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1e6)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// renderAdditionalInfo отображает дополнительную информацию
func (mp *MetricsPanel) renderAdditionalInfo() string {
	info := fmt.Sprintf(`Total Requests: %d
Successful: %d
Failed: %d
Elapsed: %v
Remaining: %v`,
		mp.metrics.TotalRequests,
		mp.metrics.SuccessfulRequests,
		mp.metrics.FailedRequests,
		mp.metrics.ElapsedTime.Round(time.Second),
		mp.metrics.RemainingTime.Round(time.Second))

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(info)
}

// Update обновляет панель метрик
func (mp *MetricsPanel) Update(msg tea.Msg) Component {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			mp.showDetails = !mp.showDetails
		case "r":
			// Refresh - обновляем все данные
			mp.updateTableData()
		}
	}

	// Обновляем прогресс бары
	mp.updateProgressBars()

	return mp
}

// GetMetrics возвращает текущие метрики
func (mp *MetricsPanel) GetMetrics() *Metrics {
	return mp.metrics
}
