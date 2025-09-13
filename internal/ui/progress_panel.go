package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressPanel отображает прогресс выполнения теста с детальной информацией
type ProgressPanel struct {
	BaseComponent
	metrics *Metrics

	// Компоненты
	mainProgress    progress.Model
	timeProgress    progress.Model
	statusIndicator *StatusIndicator

	// Состояние
	showDetails bool
	lastUpdate  time.Time
}

// StatusIndicator отображает индикатор статуса теста
type StatusIndicator struct {
	status TestStatus
	blink  bool
}

// NewProgressPanel создает новую панель прогресса
func NewProgressPanel(metrics *Metrics) *ProgressPanel {
	pp := &ProgressPanel{
		BaseComponent: BaseComponent{},
		metrics:       metrics,
		showDetails:   false,
		lastUpdate:    time.Now(),
	}

	pp.initComponents()
	return pp
}

// initComponents инициализирует компоненты
func (pp *ProgressPanel) initComponents() {
	// Основной прогресс бар
	pp.mainProgress = progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	pp.mainProgress.Width = 50

	// Прогресс времени (для визуализации времени выполнения)
	pp.timeProgress = progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	pp.timeProgress.Width = 30

	// Индикатор статуса
	pp.statusIndicator = &StatusIndicator{
		status: StatusRunning,
		blink:  false,
	}
}

// Render отображает панель прогресса
func (pp *ProgressPanel) Render(width, height int) string {
	pp.SetSize(width, height)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Render("Test Progress")

	// Обновляем прогресс
	pp.updateProgress()

	// Основной прогресс
	progressPercent := pp.metrics.GetProgress()
	progressBar := pp.mainProgress.View()
	progressText := fmt.Sprintf("%.1f%%", progressPercent*100)

	// Статус и время
	statusText := pp.renderStatus()
	timeInfo := pp.renderTimeInfo()

	// Дополнительная информация
	additionalInfo := pp.renderAdditionalInfo()

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Press 'd' for details | 's' to stop | 'p' to pause")

	content := fmt.Sprintf(`%s

%s
%s %s

%s

%s

%s`,
		title,
		progressText, progressBar,
		statusText,
		timeInfo,
		additionalInfo,
		helpText)

	return PanelStyle.Render(content)
}

// updateProgress обновляет прогресс бары
func (pp *ProgressPanel) updateProgress() {
	// Основной прогресс (общий прогресс теста)
	progressPercent := pp.metrics.GetProgress()
	pp.mainProgress.SetPercent(progressPercent)

	// Прогресс времени (сколько времени прошло относительно общей длительности)
	if pp.metrics.config != nil && pp.metrics.config.Test.Duration > 0 {
		timeProgress := float64(pp.metrics.ElapsedTime) / float64(pp.metrics.config.Test.Duration)
		if timeProgress > 1.0 {
			timeProgress = 1.0
		}
		pp.timeProgress.SetPercent(timeProgress)
	}

	// Обновляем статус
	pp.statusIndicator.status = pp.metrics.GetStatus()

	// Мигание для активного статуса
	if pp.statusIndicator.status == StatusRunning {
		pp.statusIndicator.blink = time.Since(pp.lastUpdate).Milliseconds()%1000 < 500
	} else {
		pp.statusIndicator.blink = false
	}

	pp.lastUpdate = time.Now()
}

// renderStatus отображает статус теста
func (pp *ProgressPanel) renderStatus() string {
	status := pp.statusIndicator.status
	statusText := string(status)

	var style lipgloss.Style
	switch status {
	case StatusRunning:
		if pp.statusIndicator.blink {
			style = SuccessStyle.Bold(true)
		} else {
			style = SuccessStyle
		}
		statusText = "▶ " + statusText
	case StatusPaused:
		style = WarningStyle.Bold(true)
		statusText = "⏸ " + statusText
	case StatusStopped:
		style = ErrorStyle.Bold(true)
		statusText = "⏹ " + statusText
	case StatusFinished:
		style = InfoStyle.Bold(true)
		statusText = "✅ " + statusText
	case StatusError:
		style = ErrorStyle.Bold(true)
		statusText = "❌ " + statusText
	default:
		style = MutedStyle
		statusText = "● " + statusText
	}

	return style.Render(statusText)
}

// renderTimeInfo отображает информацию о времени
func (pp *ProgressPanel) renderTimeInfo() string {
	elapsed := pp.metrics.ElapsedTime.Round(time.Second)
	remaining := pp.metrics.RemainingTime.Round(time.Second)

	// Форматируем время
	elapsedStr := pp.formatDuration(elapsed)
	remainingStr := pp.formatDuration(remaining)

	// Прогресс времени
	timeProgress := pp.timeProgress.View()

	// Скорость выполнения
	speed := pp.calculateSpeed()
	speedText := fmt.Sprintf("Speed: %.1fx", speed)

	// ETA (Estimated Time of Arrival)
	eta := pp.calculateETA()
	etaText := ""
	if eta > 0 {
		etaText = fmt.Sprintf("ETA: %s", pp.formatDuration(eta))
	}

	timeInfo := fmt.Sprintf(`Elapsed: %s | Remaining: %s
%s
%s %s`,
		elapsedStr, remainingStr,
		timeProgress,
		speedText, etaText)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(timeInfo)
}

// renderAdditionalInfo отображает дополнительную информацию
func (pp *ProgressPanel) renderAdditionalInfo() string {
	if !pp.showDetails {
		return ""
	}

	// Детальная информация о прогрессе
	details := fmt.Sprintf(`Detailed Progress:
  Requests: %d completed
  Rate: %.1f RPS (target: %d)
  Success: %.1f%% (%d/%d)
  Errors: %d (%.1f%%)
  Throughput: %.2f MB/s`,
		pp.metrics.TotalRequests,
		pp.metrics.CurrentRPS,
		pp.metrics.TargetRPS,
		pp.metrics.SuccessRate,
		pp.metrics.SuccessfulRequests,
		pp.metrics.TotalRequests,
		pp.metrics.FailedRequests,
		pp.metrics.ErrorRate,
		pp.metrics.ThroughputMBps)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Render(details)
}

// formatDuration форматирует длительность
func (pp *ProgressPanel) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-d.Minutes()*60)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - hours*60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// calculateSpeed вычисляет скорость выполнения теста
func (pp *ProgressPanel) calculateSpeed() float64 {
	if pp.metrics.ElapsedTime.Seconds() == 0 {
		return 1.0
	}

	// Скорость относительно целевого RPS
	if pp.metrics.TargetRPS > 0 {
		return pp.metrics.CurrentRPS / float64(pp.metrics.TargetRPS)
	}

	return 1.0
}

// calculateETA вычисляет оставшееся время до завершения
func (pp *ProgressPanel) calculateETA() time.Duration {
	if pp.metrics.CurrentRPS == 0 {
		return 0
	}

	// Если тест по времени
	if pp.metrics.config != nil && pp.metrics.config.Test.Duration > 0 {
		return pp.metrics.RemainingTime
	}

	// Если тест по количеству запросов (если бы была такая опция)
	// Пока возвращаем 0
	return 0
}

// Update обновляет панель прогресса
func (pp *ProgressPanel) Update(msg tea.Msg) Component {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			pp.showDetails = !pp.showDetails
		case "s":
			// Stop test
			pp.metrics.SetStatus(StatusStopped)
		case "p":
			// Pause/Resume test
			if pp.metrics.GetStatus() == StatusRunning {
				pp.metrics.SetStatus(StatusPaused)
			} else if pp.metrics.GetStatus() == StatusPaused {
				pp.metrics.SetStatus(StatusRunning)
			}
		}
	}

	// Обновляем прогресс
	pp.updateProgress()

	return pp
}

// GetProgress возвращает текущий прогресс (0.0 - 1.0)
func (pp *ProgressPanel) GetProgress() float64 {
	return pp.metrics.GetProgress()
}

// GetStatus возвращает текущий статус
func (pp *ProgressPanel) GetStatus() TestStatus {
	return pp.metrics.GetStatus()
}

// SetStatus устанавливает статус
func (pp *ProgressPanel) SetStatus(status TestStatus) {
	pp.metrics.SetStatus(status)
}
