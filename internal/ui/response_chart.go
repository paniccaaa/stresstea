package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResponseTimeChart отображает график времени отклика в реальном времени
type ResponseTimeChart struct {
	BaseComponent
	metrics *Metrics

	// Данные графика
	responseData []ResponseDataPoint
	rpsData      []RPSDataPoint

	// Настройки отображения
	chartWidth      int
	chartHeight     int
	timeWindow      time.Duration // Окно времени для отображения
	showRPS         bool
	showPercentiles bool

	// Состояние
	selectedMetric string // "response", "rps", "percentiles"
	animationFrame int
}

// ResponseDataPoint представляет точку данных времени отклика
type ResponseDataPoint struct {
	Timestamp time.Time
	Latency   time.Duration
	Status    int
	Error     bool
}

// RPSDataPoint представляет точку данных RPS
type RPSDataPoint struct {
	Timestamp time.Time
	RPS       float64
}

// PercentileData содержит данные о перцентилях
type PercentileData struct {
	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
}

// NewResponseTimeChart создает новый график времени отклика
func NewResponseTimeChart(metrics *Metrics) *ResponseTimeChart {
	rc := &ResponseTimeChart{
		BaseComponent:   BaseComponent{},
		metrics:         metrics,
		chartWidth:      40,
		chartHeight:     12,
		timeWindow:      60 * time.Second, // 60 секунд
		showRPS:         false,
		showPercentiles: true,
		selectedMetric:  "response",
		animationFrame:  0,
	}

	return rc
}

// Render отображает график времени отклика
func (rc *ResponseTimeChart) Render(width, height int) string {
	rc.SetSize(width, height)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Render("Response Time Chart")

	// Обновляем данные
	rc.updateData()

	// Основной график
	chartContent := rc.renderMainChart()

	// Легенда
	legend := rc.renderLegend()

	// Статистика
	stats := rc.renderStats()

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Press 'm' for metrics | 'r' for RPS | 'p' for percentiles | 'z' to zoom")

	content := fmt.Sprintf(`%s

%s

%s

%s

%s`,
		title,
		chartContent,
		legend,
		stats,
		helpText)

	return PanelStyle.Render(content)
}

// updateData обновляет данные графика
func (rc *ResponseTimeChart) updateData() {
	// Обновляем данные времени отклика
	rc.updateResponseData()

	// Обновляем данные RPS
	rc.updateRPSData()
}

// updateResponseData обновляет данные времени отклика
func (rc *ResponseTimeChart) updateResponseData() {
	// Очищаем старые данные
	cutoff := time.Now().Add(-rc.timeWindow)
	var newData []ResponseDataPoint

	for _, point := range rc.responseData {
		if point.Timestamp.After(cutoff) {
			newData = append(newData, point)
		}
	}

	rc.responseData = newData
}

// updateRPSData обновляет данные RPS
func (rc *ResponseTimeChart) updateRPSData() {
	// Используем историю RPS из метрик
	if len(rc.metrics.RPSHistory) > 0 {
		rc.rpsData = make([]RPSDataPoint, len(rc.metrics.RPSHistory))

		now := time.Now()
		for i, rps := range rc.metrics.RPSHistory {
			rc.rpsData[i] = RPSDataPoint{
				Timestamp: now.Add(-time.Duration(len(rc.metrics.RPSHistory)-i) * time.Second),
				RPS:       rps,
			}
		}
	}
}

// renderMainChart отображает основной график
func (rc *ResponseTimeChart) renderMainChart() string {
	switch rc.selectedMetric {
	case "response":
		return rc.renderResponseTimeChart()
	case "rps":
		return rc.renderRPSChart()
	case "percentiles":
		return rc.renderPercentilesChart()
	default:
		return rc.renderResponseTimeChart()
	}
}

// renderResponseTimeChart отображает график времени отклика
func (rc *ResponseTimeChart) renderResponseTimeChart() string {
	if len(rc.responseData) == 0 {
		return "No response data available"
	}

	// Находим максимальное время отклика
	maxLatency := time.Duration(0)
	for _, point := range rc.responseData {
		if point.Latency > maxLatency {
			maxLatency = point.Latency
		}
	}

	if maxLatency == 0 {
		maxLatency = 100 * time.Millisecond
	}

	// Создаем ASCII график
	chart := ""

	// Y-ось (время отклика)
	for i := rc.chartHeight - 1; i >= 0; i-- {
		// Вычисляем значение для этой строки
		value := float64(i) / float64(rc.chartHeight-1) * float64(maxLatency.Nanoseconds())
		threshold := time.Duration(value)

		// Форматируем значение
		valueStr := rc.formatLatency(threshold)

		// Создаем строку графика
		line := fmt.Sprintf("%-8s │", valueStr)

		// Добавляем точки данных
		for j := 0; j < rc.chartWidth; j++ {
			if j < len(rc.responseData) {
				point := rc.responseData[j]
				if point.Latency >= threshold {
					// Определяем символ в зависимости от статуса
					symbol := rc.getResponseSymbol(point)
					line += symbol
				} else {
					line += " "
				}
			} else {
				line += " "
			}
		}

		chart += line + "\n"
	}

	// X-ось (время)
	chart += "         └"
	for i := 0; i < rc.chartWidth; i++ {
		chart += "─"
	}

	return chart
}

// renderRPSChart отображает график RPS
func (rc *ResponseTimeChart) renderRPSChart() string {
	if len(rc.rpsData) == 0 {
		return "No RPS data available"
	}

	// Находим максимальный RPS
	maxRPS := 0.0
	for _, point := range rc.rpsData {
		if point.RPS > maxRPS {
			maxRPS = point.RPS
		}
	}

	if maxRPS == 0 {
		maxRPS = 100
	}

	// Создаем ASCII график
	chart := ""

	// Y-ось (RPS)
	for i := rc.chartHeight - 1; i >= 0; i-- {
		value := float64(i) / float64(rc.chartHeight-1) * maxRPS
		valueStr := fmt.Sprintf("%.0f", value)

		line := fmt.Sprintf("%-8s │", valueStr)

		// Добавляем точки данных
		for j := 0; j < rc.chartWidth; j++ {
			if j < len(rc.rpsData) {
				point := rc.rpsData[j]
				if point.RPS >= value {
					line += BlockChar
				} else {
					line += " "
				}
			} else {
				line += " "
			}
		}

		chart += line + "\n"
	}

	// X-ось
	chart += "         └"
	for i := 0; i < rc.chartWidth; i++ {
		chart += "─"
	}

	return chart
}

// renderPercentilesChart отображает график перцентилей
func (rc *ResponseTimeChart) renderPercentilesChart() string {
	chart := "Percentiles Chart:\n\n"

	// Получаем текущие перцентили
	p50 := rc.metrics.P50Latency
	p90 := rc.metrics.P90Latency
	p95 := rc.metrics.P95Latency
	p99 := rc.metrics.P99Latency

	// Находим максимальное значение
	maxValue := p99
	if maxValue == 0 {
		maxValue = 100 * time.Millisecond
	}

	// Создаем график перцентилей
	percentiles := []struct {
		name  string
		value time.Duration
		color lipgloss.Color
	}{
		{"P50", p50, lipgloss.Color("#00FF00")},
		{"P90", p90, lipgloss.Color("#FFA500")},
		{"P95", p95, lipgloss.Color("#FF4500")},
		{"P99", p99, lipgloss.Color("#FF0000")},
	}

	for _, p := range percentiles {
		// Вычисляем длину бара
		barLength := int(float64(p.value) / float64(maxValue) * float64(rc.chartWidth))

		// Создаем бар
		bar := ""
		for i := 0; i < rc.chartWidth; i++ {
			if i < barLength {
				bar += BarChar
			} else {
				bar += " "
			}
		}

		// Стилизуем
		style := lipgloss.NewStyle().Foreground(p.color)
		valueStr := rc.formatLatency(p.value)

		chart += fmt.Sprintf("%s: %s %s\n",
			style.Render(p.name),
			bar,
			valueStr)
	}

	return chart
}

// renderLegend отображает легенду
func (rc *ResponseTimeChart) renderLegend() string {
	legend := "Legend: "

	switch rc.selectedMetric {
	case "response":
		legend += "● Success | ▲ Warning | ▼ Error"
	case "rps":
		legend += "█ RPS (requests per second)"
	case "percentiles":
		legend += "P50: Green | P90: Orange | P95: Red | P99: Dark Red"
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(legend)
}

// renderStats отображает статистику
func (rc *ResponseTimeChart) renderStats() string {
	stats := ""

	switch rc.selectedMetric {
	case "response":
		stats = fmt.Sprintf(`Response Time Stats:
Avg: %s | Min: %s | Max: %s
P90: %s | P95: %s | P99: %s`,
			rc.formatLatency(rc.metrics.AvgLatency),
			rc.formatLatency(rc.metrics.MinLatency),
			rc.formatLatency(rc.metrics.MaxLatency),
			rc.formatLatency(rc.metrics.P90Latency),
			rc.formatLatency(rc.metrics.P95Latency),
			rc.formatLatency(rc.metrics.P99Latency))
	case "rps":
		stats = fmt.Sprintf(`RPS Stats:
Current: %.1f | Target: %d | Max: %.1f`,
			rc.metrics.CurrentRPS,
			rc.metrics.TargetRPS,
			rc.getMaxRPS())
	case "percentiles":
		stats = fmt.Sprintf(`Percentile Distribution:
P50: %s (50%% of requests)
P90: %s (90%% of requests)
P95: %s (95%% of requests)
P99: %s (99%% of requests)`,
			rc.formatLatency(rc.metrics.P50Latency),
			rc.formatLatency(rc.metrics.P90Latency),
			rc.formatLatency(rc.metrics.P95Latency),
			rc.formatLatency(rc.metrics.P99Latency))
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(stats)
}

// getResponseSymbol возвращает символ для точки данных
func (rc *ResponseTimeChart) getResponseSymbol(point ResponseDataPoint) string {
	if point.Error {
		return "▼" // Ошибка
	}

	switch {
	case point.Status >= 200 && point.Status < 300:
		return "●" // Успех
	case point.Status >= 300 && point.Status < 500:
		return "▲" // Перенаправление/ошибка клиента
	default:
		return "▼" // Ошибка сервера
	}
}

// formatLatency форматирует время отклика
func (rc *ResponseTimeChart) formatLatency(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.0fns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1e6)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// getMaxRPS возвращает максимальный RPS
func (rc *ResponseTimeChart) getMaxRPS() float64 {
	max := 0.0
	for _, point := range rc.rpsData {
		if point.RPS > max {
			max = point.RPS
		}
	}
	return max
}

// Update обновляет график времени отклика
func (rc *ResponseTimeChart) Update(msg tea.Msg) Component {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "m":
			rc.selectedMetric = "response"
		case "r":
			rc.selectedMetric = "rps"
		case "p":
			rc.selectedMetric = "percentiles"
		case "z":
			rc.cycleTimeWindow()
		}
	}

	// Обновляем анимацию
	rc.animationFrame++

	return rc
}

// cycleTimeWindow переключает между окнами времени
func (rc *ResponseTimeChart) cycleTimeWindow() {
	windows := []time.Duration{
		30 * time.Second,
		60 * time.Second,
		120 * time.Second,
		300 * time.Second,
	}

	currentIndex := 0
	for i, window := range windows {
		if window == rc.timeWindow {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(windows)
	rc.timeWindow = windows[nextIndex]
}

// AddResponseData добавляет новую точку данных времени отклика
func (rc *ResponseTimeChart) AddResponseData(latency time.Duration, status int, hasError bool) {
	point := ResponseDataPoint{
		Timestamp: time.Now(),
		Latency:   latency,
		Status:    status,
		Error:     hasError,
	}

	rc.responseData = append(rc.responseData, point)

	// Ограничиваем количество точек
	if len(rc.responseData) > 1000 {
		rc.responseData = rc.responseData[1:]
	}
}
