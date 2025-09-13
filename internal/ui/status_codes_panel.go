package ui

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusCodesPanel отображает распределение статус кодов с таблицей и графиком
type StatusCodesPanel struct {
	BaseComponent
	metrics *Metrics

	// Компоненты
	statusTable table.Model
	barChart    *BarChart

	// Состояние
	showChart   bool
	sortBy      string // "count", "percentage", "status"
	sortOrder   string // "asc", "desc"
	selectedRow int
}

// BarChart представляет ASCII бар-чарт
type BarChart struct {
	data     []StatusCodeData
	maxValue int
	width    int
}

// StatusCodeData содержит данные о статус коде
type StatusCodeData struct {
	Status     int
	Count      int
	Percentage float64
	Color      lipgloss.Color
}

// NewStatusCodesPanel создает новую панель статус кодов
func NewStatusCodesPanel(metrics *Metrics) *StatusCodesPanel {
	scp := &StatusCodesPanel{
		BaseComponent: BaseComponent{},
		metrics:       metrics,
		showChart:     true,
		sortBy:        "count",
		sortOrder:     "desc",
		selectedRow:   0,
	}

	scp.initComponents()
	return scp
}

// initComponents инициализирует компоненты
func (scp *StatusCodesPanel) initComponents() {
	// Таблица статус кодов
	scp.initStatusTable()

	// Бар-чарт
	scp.barChart = &BarChart{
		data:  []StatusCodeData{},
		width: 20,
	}
}

// initStatusTable инициализирует таблицу статус кодов
func (scp *StatusCodesPanel) initStatusTable() {
	columns := []table.Column{
		{Title: "Status", Width: 8},
		{Title: "Count", Width: 10},
		{Title: "Percent", Width: 10},
		{Title: "Chart", Width: 20},
	}

	rows := []table.Row{
		{"200", "0", "0.0%", ""},
		{"404", "0", "0.0%", ""},
		{"500", "0", "0.0%", ""},
	}

	scp.statusTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(6),
	)

	// Стилизация таблицы
	scp.statusTable.SetStyles(table.Styles{
		Header: TableHeaderStyle,
		Cell:   TableRowStyle,
	})
}

// Render отображает панель статус кодов
func (scp *StatusCodesPanel) Render(width, height int) string {
	scp.SetSize(width, height)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Render("Status Codes Distribution")

	// Обновляем данные
	scp.updateData()

	// Таблица
	tableContent := scp.renderTable()

	// Бар-чарт (если включен)
	chartContent := ""
	if scp.showChart {
		chartContent = scp.renderBarChart()
	}

	// Сводка
	summary := scp.renderSummary()

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Press 'c' for chart | 's' to sort | 'r' to refresh")

	content := fmt.Sprintf(`%s

%s

%s

%s

%s`,
		title,
		tableContent,
		chartContent,
		summary,
		helpText)

	return PanelStyle.Render(content)
}

// updateData обновляет данные таблицы и графика
func (scp *StatusCodesPanel) updateData() {
	// Получаем данные о статус кодах
	statusCodes := scp.metrics.GetStatusCodesSorted()

	// Создаем данные для таблицы
	var rows []table.Row
	var chartData []StatusCodeData

	// Ограничиваем количество отображаемых статус кодов
	maxRows := 8
	if len(statusCodes) > maxRows {
		statusCodes = statusCodes[:maxRows]
	}

	for _, info := range statusCodes {
		// Строка таблицы
		statusStr := strconv.Itoa(info.Status)
		countStr := strconv.Itoa(info.Count)
		percentStr := fmt.Sprintf("%.1f%%", info.Percentage)
		chartBar := scp.renderStatusBar(info.Percentage, 15)

		rows = append(rows, table.Row{statusStr, countStr, percentStr, chartBar})

		// Данные для графика
		color := scp.getStatusColor(info.Status)
		chartData = append(chartData, StatusCodeData{
			Status:     info.Status,
			Count:      info.Count,
			Percentage: info.Percentage,
			Color:      color,
		})
	}

	// Обновляем таблицу
	scp.statusTable.SetRows(rows)

	// Обновляем график
	scp.barChart.data = chartData
	if len(chartData) > 0 {
		scp.barChart.maxValue = chartData[0].Count
	}
}

// renderTable отображает таблицу статус кодов
func (scp *StatusCodesPanel) renderTable() string {
	tableContent := scp.statusTable.View()

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Render(tableContent)
}

// renderBarChart отображает ASCII бар-чарт
func (scp *StatusCodesPanel) renderBarChart() string {
	if len(scp.barChart.data) == 0 {
		return ""
	}

	chart := "Status Codes Chart:\n"

	for _, data := range scp.barChart.data {
		// Вычисляем длину бара
		barLength := 0
		if scp.barChart.maxValue > 0 {
			barLength = int(float64(data.Count) / float64(scp.barChart.maxValue) * float64(scp.barChart.width))
		}

		// Создаем бар
		bar := ""
		for i := 0; i < scp.barChart.width; i++ {
			if i < barLength {
				bar += BarChar
			} else {
				bar += " "
			}
		}

		// Стилизуем строку
		statusStr := fmt.Sprintf("%d", data.Status)
		style := lipgloss.NewStyle().Foreground(data.Color)

		chart += fmt.Sprintf("%s: %s %d (%.1f%%)\n",
			style.Render(statusStr),
			bar,
			data.Count,
			data.Percentage)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Render(chart)
}

// renderStatusBar создает ASCII бар для статус кода
func (scp *StatusCodesPanel) renderStatusBar(percentage float64, width int) string {
	filled := int(percentage * float64(width) / 100)
	bar := ""

	for i := 0; i < width; i++ {
		if i < filled {
			bar += BarChar
		} else {
			bar += " "
		}
	}

	return bar
}

// renderSummary отображает сводку по статус кодам
func (scp *StatusCodesPanel) renderSummary() string {
	total := scp.metrics.TotalRequests
	successful := scp.metrics.SuccessfulRequests
	failed := scp.metrics.FailedRequests

	successRate := 0.0
	errorRate := 0.0

	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
		errorRate = float64(failed) / float64(total) * 100
	}

	summary := fmt.Sprintf(`Summary:
Total Requests: %d
Successful: %d (%.1f%%)
Failed: %d (%.1f%%)`,
		total,
		successful, successRate,
		failed, errorRate)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(summary)
}

// getStatusColor возвращает цвет для статус кода
func (scp *StatusCodesPanel) getStatusColor(status int) lipgloss.Color {
	switch {
	case status >= 200 && status < 300:
		return lipgloss.Color("#00FF00") // Зеленый
	case status >= 300 && status < 400:
		return lipgloss.Color("#00BFFF") // Голубой
	case status >= 400 && status < 500:
		return lipgloss.Color("#FFA500") // Оранжевый
	case status >= 500:
		return lipgloss.Color("#FF0000") // Красный
	default:
		return lipgloss.Color("#626262") // Серый
	}
}

// Update обновляет панель статус кодов
func (scp *StatusCodesPanel) Update(msg tea.Msg) Component {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "c":
			scp.showChart = !scp.showChart
		case "s":
			scp.cycleSort()
		case "r":
			// Refresh - обновляем данные
			scp.updateData()
		case KeyUp:
			if scp.selectedRow > 0 {
				scp.selectedRow--
			}
		case KeyDown:
			if scp.selectedRow < len(scp.statusTable.Rows())-1 {
				scp.selectedRow++
			}
		}
	}

	return scp
}

// cycleSort переключает между вариантами сортировки
func (scp *StatusCodesPanel) cycleSort() {
	sortOptions := []string{"count", "percentage", "status"}
	currentIndex := 0

	for i, option := range sortOptions {
		if option == scp.sortBy {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(sortOptions)
	scp.sortBy = sortOptions[nextIndex]

	// Применяем сортировку
	scp.applySorting()
}

// applySorting применяет текущую сортировку
func (scp *StatusCodesPanel) applySorting() {
	statusCodes := scp.metrics.GetStatusCodesSorted()

	switch scp.sortBy {
	case "count":
		sort.Slice(statusCodes, func(i, j int) bool {
			if scp.sortOrder == "asc" {
				return statusCodes[i].Count < statusCodes[j].Count
			}
			return statusCodes[i].Count > statusCodes[j].Count
		})
	case "percentage":
		sort.Slice(statusCodes, func(i, j int) bool {
			if scp.sortOrder == "asc" {
				return statusCodes[i].Percentage < statusCodes[j].Percentage
			}
			return statusCodes[i].Percentage > statusCodes[j].Percentage
		})
	case "status":
		sort.Slice(statusCodes, func(i, j int) bool {
			if scp.sortOrder == "asc" {
				return statusCodes[i].Status < statusCodes[j].Status
			}
			return statusCodes[i].Status > statusCodes[j].Status
		})
	}

	// Обновляем данные с новой сортировкой
	scp.updateData()
}

// GetStatusCodes возвращает данные о статус кодах
func (scp *StatusCodesPanel) GetStatusCodes() []StatusCodeInfo {
	return scp.metrics.GetStatusCodesSorted()
}
