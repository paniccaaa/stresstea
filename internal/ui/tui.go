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

// TUI представляет основной интерфейс приложения
type TUI struct {
	config  *parser.Config
	results []loadtest.Result
	start   time.Time
}

// model представляет состояние TUI приложения
type model struct {
	config      *parser.Config
	results     []loadtest.Result
	metrics     *Metrics
	start       time.Time
	width       int
	height      int
	resultsChan chan loadtest.Result

	// Компоненты UI
	configPanel      *ConfigPanel
	metricsPanel     *MetricsPanel
	progressPanel    *ProgressPanel
	statusCodesPanel *StatusCodesPanel
	errorLogPanel    *ErrorLogPanel
	responseChart    *ResponseTimeChart

	// Состояние
	status       TestStatus
	showHelp     bool
	focusedPanel string
}

// NewTUI создает новый экземпляр TUI
func NewTUI(cfg *parser.Config) *TUI {
	return &TUI{
		config: cfg,
		start:  time.Now(),
	}
}

// Run запускает TUI приложение
func (t *TUI) Run(resultsChan chan loadtest.Result) error {
	metrics := NewMetrics(t.config)

	m := model{
		config:       t.config,
		metrics:      metrics,
		start:        t.start,
		resultsChan:  resultsChan,
		status:       StatusRunning,
		focusedPanel: "config",
	}

	// Инициализируем компоненты
	m.configPanel = NewConfigPanel(t.config)
	m.metricsPanel = NewMetricsPanel(metrics)
	m.progressPanel = NewProgressPanel(metrics)
	m.statusCodesPanel = NewStatusCodesPanel(metrics)
	m.errorLogPanel = NewErrorLogPanel(metrics)
	m.responseChart = NewResponseTimeChart(metrics)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// UpdateResults обновляет результаты (deprecated, используется resultsChan)
func (t *TUI) UpdateResults(results chan loadtest.Result) {
	for result := range results {
		t.results = append(t.results, result)
	}
}

// Init инициализирует модель
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.waitForResults(),
	)
}

// waitForResults ожидает новые результаты
func (m model) waitForResults() tea.Cmd {
	return func() tea.Msg {
		if m.resultsChan != nil {
			select {
			case result := <-m.resultsChan:
				return result
			default:
				// Return a tick to continue polling
				return tea.Tick(time.Millisecond*UpdateInterval, func(t time.Time) tea.Msg {
					return t
				})()
			}
		}
		return nil
	}
}

// Update обновляет состояние модели
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyQuit, KeyCtrlC:
			return m, tea.Quit
		case KeyHelp:
			m.showHelp = !m.showHelp
			return m, nil
		case KeyPause:
			if m.status == StatusRunning {
				m.status = StatusPaused
			} else if m.status == StatusPaused {
				m.status = StatusRunning
			}
			return m, nil
		case KeyStop:
			m.status = StatusStopped
			return m, nil
		case KeyTab:
			m.switchPanel()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Обновляем размеры компонентов
		m.updateComponentSizes()

	case loadtest.Result:
		if m.status == StatusRunning {
			m.results = append(m.results, msg)
			// Ограничиваем количество результатов в памяти
			if len(m.results) > MaxResults {
				m.results = m.results[len(m.results)-MaxResults:]
			}
			// Обновляем метрики
			m.metrics.UpdateMetrics(m.results)
		}
		return m, m.waitForResults()
	case time.Time:
		// Tick event, continue polling
		return m, m.waitForResults()
	}

	return m, nil
}

// View отображает интерфейс
func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Проверяем минимальный размер терминала
	if m.width < MinWidth || m.height < MinHeight {
		return fmt.Sprintf("%s\nMinimum terminal size: %dx%d\nCurrent size: %dx%d",
			ErrTerminalTooSmall, MinWidth, MinHeight, m.width, m.height)
	}

	// Показываем help если нужно
	if m.showHelp {
		return m.renderHelp()
	}

	// Основной интерфейс
	content := m.renderMainInterface()

	// Ограничиваем высоту контента
	lines := strings.Split(content, "\n")
	if len(lines) > m.height-2 {
		lines = lines[:m.height-2]
		content = strings.Join(lines, "\n")
	}

	return content
}

// renderMainInterface отображает основной интерфейс
func (m model) renderMainInterface() string {
	// Заголовок
	title := lipgloss.Place(
		m.width, 1,
		lipgloss.Center, lipgloss.Center,
		TitleStyle.Render("🚀 Stresstea - HTTP Load Testing"),
	)

	// Рассчитываем доступную высоту для панелей
	availableHeight := m.height - 4 // -4 для заголовка, помощи и отступов

	// Верхняя панель (конфигурация и метрики) - 40% от доступной высоты
	topHeight := int(float64(availableHeight) * 0.4)
	topPanel := m.renderTopPanelWithHeight(topHeight)

	// Панель прогресса - 15% от доступной высоты
	progressHeight := int(float64(availableHeight) * 0.15)
	progressPanel := m.renderProgressPanelWithHeight(progressHeight)

	// Нижняя панель - 45% от доступной высоты
	bottomHeight := availableHeight - topHeight - progressHeight
	bottomPanel := m.renderBottomPanelWithHeight(bottomHeight)

	// Помощь
	help := lipgloss.Place(
		m.width, 1,
		lipgloss.Center, lipgloss.Center,
		HelpStyle.Render("Press 'h' for help, 'q' to quit"),
	)

	// Собираем все вместе с минимальными отступами
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		topPanel,
		"",
		progressPanel,
		"",
		bottomPanel,
		"",
		help,
	)

	return content
}

// renderTopPanelWithHeight отображает верхнюю панель с заданной высотой
func (m model) renderTopPanelWithHeight(height int) string {
	// Адаптивные размеры с учетом границ панелей
	configWidth := (m.width - 8) / 2 // -8 для границ и отступов
	metricsWidth := (m.width - 8) / 2

	configPanel := m.configPanel.Render(configWidth, height)
	metricsPanel := m.metricsPanel.Render(metricsWidth, height)

	// Создаем отступ между панелями
	spacer := lipgloss.NewStyle().Width(2).Render("")

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		configPanel,
		spacer,
		metricsPanel,
	)
}

// renderProgressPanelWithHeight отображает панель прогресса с заданной высотой
func (m model) renderProgressPanelWithHeight(height int) string {
	return m.progressPanel.Render(m.width, height)
}

// renderBottomPanelWithHeight отображает нижнюю панель с заданной высотой
func (m model) renderBottomPanelWithHeight(height int) string {
	// Адаптивные размеры с учетом границ панелей
	panelWidth := (m.width - 12) / 3 // -12 для границ и отступов

	statusPanel := m.statusCodesPanel.Render(panelWidth, height)
	chartPanel := m.responseChart.Render(panelWidth, height)
	errorPanel := m.errorLogPanel.Render(panelWidth, height)

	// Создаем отступы между панелями
	spacer := lipgloss.NewStyle().Width(2).Render("")

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		statusPanel,
		spacer,
		chartPanel,
		spacer,
		errorPanel,
	)
}

// renderHelp отображает справку
func (m model) renderHelp() string {
	return fmt.Sprintf("%s\n\n%s\n\nPress 'h' to close help",
		TitleStyle.Render("Help"), HelpMessage)
}

// switchPanel переключает между панелями
func (m *model) switchPanel() {
	panels := []string{"config", "stats", "progress", "status", "chart", "errors"}
	currentIndex := 0

	for i, panel := range panels {
		if panel == m.focusedPanel {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(panels)
	m.focusedPanel = panels[nextIndex]
}

// updateComponentSizes обновляет размеры компонентов
func (m *model) updateComponentSizes() {
	// Обновляем размеры всех компонентов
	if m.configPanel != nil {
		m.configPanel.SetSize(m.width/2, ConfigPanelHeight)
	}
	if m.metricsPanel != nil {
		m.metricsPanel.SetSize(m.width-m.width/2-1, MetricsPanelHeight)
	}
	if m.progressPanel != nil {
		m.progressPanel.SetSize(m.width, ProgressPanelHeight)
	}
	if m.statusCodesPanel != nil {
		m.statusCodesPanel.SetSize(m.width/3, StatusCodesPanelHeight)
	}
	if m.responseChart != nil {
		m.responseChart.SetSize(m.width/3, ResponseChartHeight)
	}
	if m.errorLogPanel != nil {
		m.errorLogPanel.SetSize(m.width/3, ErrorLogPanelHeight)
	}
}
