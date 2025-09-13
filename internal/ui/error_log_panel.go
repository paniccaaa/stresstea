package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorLogPanel отображает прокручиваемый лог ошибок
type ErrorLogPanel struct {
	BaseComponent
	metrics *Metrics

	// Данные лога
	errors []LogEntry

	// Состояние прокрутки
	scrollOffset int
	maxLines     int

	// Фильтрация
	filterText    string
	filterLevel   string // "all", "error", "warning", "info"
	showTimestamp bool

	// Состояние
	selectedIndex int
	showDetails   bool
}

// LogEntry представляет запись в логе
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Details   string
	Count     int // Количество повторений
}

// LogLevel представляет уровень лога
type LogLevel string

const (
	LogLevelError   LogLevel = "ERROR"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelDebug   LogLevel = "DEBUG"
)

// NewErrorLogPanel создает новую панель лога ошибок
func NewErrorLogPanel(metrics *Metrics) *ErrorLogPanel {
	elp := &ErrorLogPanel{
		BaseComponent: BaseComponent{},
		metrics:       metrics,
		errors:        []LogEntry{},
		scrollOffset:  0,
		maxLines:      10,
		filterText:    "",
		filterLevel:   "all",
		showTimestamp: true,
		selectedIndex: 0,
		showDetails:   false,
	}

	return elp
}

// Render отображает панель лога ошибок
func (elp *ErrorLogPanel) Render(width, height int) string {
	elp.SetSize(width, height)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Render("Error Log")

	// Обновляем данные
	elp.updateErrors()

	// Фильтруем ошибки
	filteredErrors := elp.filterErrors()

	// Лог ошибок
	logContent := elp.renderLog(filteredErrors)

	// Статистика
	stats := elp.renderStats()

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("↑↓ Navigate | 'f' filter | 'd' details | 'c' clear | 'r' refresh")

	content := fmt.Sprintf(`%s

%s

%s

%s`,
		title,
		logContent,
		stats,
		helpText)

	return PanelStyle.Render(content)
}

// updateErrors обновляет список ошибок из метрик
func (elp *ErrorLogPanel) updateErrors() {
	// Получаем ошибки из метрик
	recentErrors := elp.metrics.GetTopErrors()

	// Преобразуем в LogEntry
	for _, errorMsg := range recentErrors {
		if errorMsg == "No errors" {
			continue
		}

		// Парсим сообщение об ошибке
		level := elp.parseErrorLevel(errorMsg)
		message := elp.parseErrorMessage(errorMsg)

		// Проверяем, есть ли уже такая ошибка
		found := false
		for i, entry := range elp.errors {
			if entry.Message == message {
				elp.errors[i].Count++
				elp.errors[i].Timestamp = time.Now()
				found = true
				break
			}
		}

		if !found {
			entry := LogEntry{
				Timestamp: time.Now(),
				Level:     level,
				Message:   message,
				Details:   errorMsg,
				Count:     1,
			}
			elp.errors = append(elp.errors, entry)
		}
	}

	// Сортируем по времени (новые сверху)
	elp.sortErrors()

	// Ограничиваем количество ошибок
	if len(elp.errors) > MaxErrors {
		elp.errors = elp.errors[:MaxErrors]
	}
}

// parseErrorLevel определяет уровень ошибки по сообщению
func (elp *ErrorLogPanel) parseErrorLevel(message string) LogLevel {
	message = strings.ToUpper(message)

	if strings.Contains(message, "TIMEOUT") || strings.Contains(message, "CONNECTION REFUSED") {
		return LogLevelError
	} else if strings.Contains(message, "WARNING") {
		return LogLevelWarning
	} else if strings.Contains(message, "INFO") {
		return LogLevelInfo
	} else {
		return LogLevelError
	}
}

// parseErrorMessage извлекает основное сообщение об ошибке
func (elp *ErrorLogPanel) parseErrorMessage(message string) string {
	// Убираем временные метки и лишние символы
	parts := strings.Split(message, " - ")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}

	return strings.TrimSpace(message)
}

// sortErrors сортирует ошибки по времени
func (elp *ErrorLogPanel) sortErrors() {
	// Сортируем по времени (новые сверху)
	for i := 0; i < len(elp.errors)-1; i++ {
		for j := i + 1; j < len(elp.errors); j++ {
			if elp.errors[i].Timestamp.Before(elp.errors[j].Timestamp) {
				elp.errors[i], elp.errors[j] = elp.errors[j], elp.errors[i]
			}
		}
	}
}

// filterErrors фильтрует ошибки по тексту и уровню
func (elp *ErrorLogPanel) filterErrors() []LogEntry {
	var filtered []LogEntry

	for _, entry := range elp.errors {
		// Фильтр по уровню
		if elp.filterLevel != "all" && string(entry.Level) != elp.filterLevel {
			continue
		}

		// Фильтр по тексту
		if elp.filterText != "" {
			message := strings.ToLower(entry.Message)
			filter := strings.ToLower(elp.filterText)
			if !strings.Contains(message, filter) {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// renderLog отображает лог ошибок
func (elp *ErrorLogPanel) renderLog(errors []LogEntry) string {
	if len(errors) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			Render("No errors to display")
	}

	// Вычисляем диапазон отображения
	start := elp.scrollOffset
	end := start + elp.maxLines
	if end > len(errors) {
		end = len(errors)
	}

	if start >= len(errors) {
		start = len(errors) - elp.maxLines
		if start < 0 {
			start = 0
		}
	}

	logContent := ""
	for i := start; i < end; i++ {
		entry := errors[i]

		// Выделяем выбранную строку
		style := elp.getEntryStyle(entry)
		if i == elp.selectedIndex {
			style = style.Background(lipgloss.Color("#3C3C3C"))
		}

		// Форматируем запись
		line := elp.formatLogEntry(entry, i == elp.selectedIndex)
		logContent += style.Render(line) + "\n"
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Render(logContent)
}

// getEntryStyle возвращает стиль для записи лога
func (elp *ErrorLogPanel) getEntryStyle(entry LogEntry) lipgloss.Style {
	switch entry.Level {
	case LogLevelError:
		return ErrorStyle
	case LogLevelWarning:
		return WarningStyle
	case LogLevelInfo:
		return InfoStyle
	case LogLevelDebug:
		return MutedStyle
	default:
		return MutedStyle
	}
}

// formatLogEntry форматирует запись лога
func (elp *ErrorLogPanel) formatLogEntry(entry LogEntry, selected bool) string {
	// Временная метка
	timestamp := ""
	if elp.showTimestamp {
		timestamp = entry.Timestamp.Format("15:04:05") + " "
	}

	// Уровень
	level := fmt.Sprintf("[%s] ", entry.Level)

	// Сообщение
	message := entry.Message

	// Счетчик повторений
	count := ""
	if entry.Count > 1 {
		count = fmt.Sprintf(" (%dx)", entry.Count)
	}

	// Обрезаем длинные сообщения
	maxLength := 50
	if len(message) > maxLength {
		message = message[:maxLength] + "..."
	}

	line := fmt.Sprintf("%s%s%s%s", timestamp, level, message, count)

	// Добавляем детали для выбранной записи
	if selected && elp.showDetails && entry.Details != "" {
		line += "\n  " + entry.Details
	}

	return line
}

// renderStats отображает статистику ошибок
func (elp *ErrorLogPanel) renderStats() string {
	totalErrors := len(elp.errors)
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, entry := range elp.errors {
		switch entry.Level {
		case LogLevelError:
			errorCount++
		case LogLevelWarning:
			warningCount++
		case LogLevelInfo:
			infoCount++
		}
	}

	stats := fmt.Sprintf(`Error Statistics:
Total: %d | Errors: %d | Warnings: %d | Info: %d`,
		totalErrors, errorCount, warningCount, infoCount)

	// Добавляем информацию о фильтрах
	if elp.filterText != "" || elp.filterLevel != "all" {
		filterInfo := fmt.Sprintf("\nFilter: %s | Level: %s",
			elp.filterText, elp.filterLevel)
		stats += filterInfo
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(stats)
}

// Update обновляет панель лога ошибок
func (elp *ErrorLogPanel) Update(msg tea.Msg) Component {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyUp:
			if elp.selectedIndex > 0 {
				elp.selectedIndex--
			}
		case KeyDown:
			filteredErrors := elp.filterErrors()
			if elp.selectedIndex < len(filteredErrors)-1 {
				elp.selectedIndex++
			}
		case "f":
			// Toggle filter (в реальном приложении здесь был бы ввод)
			if elp.filterLevel == "all" {
				elp.filterLevel = "error"
			} else if elp.filterLevel == "error" {
				elp.filterLevel = "warning"
			} else if elp.filterLevel == "warning" {
				elp.filterLevel = "info"
			} else {
				elp.filterLevel = "all"
			}
		case "d":
			elp.showDetails = !elp.showDetails
		case "c":
			elp.clearErrors()
		case "r":
			elp.updateErrors()
		}
	}

	return elp
}

// clearErrors очищает лог ошибок
func (elp *ErrorLogPanel) clearErrors() {
	elp.errors = []LogEntry{}
	elp.selectedIndex = 0
	elp.scrollOffset = 0
}

// AddError добавляет новую ошибку в лог
func (elp *ErrorLogPanel) AddError(level LogLevel, message string, details string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Details:   details,
		Count:     1,
	}

	elp.errors = append(elp.errors, entry)
	elp.sortErrors()

	// Ограничиваем количество ошибок
	if len(elp.errors) > MaxErrors {
		elp.errors = elp.errors[:MaxErrors]
	}
}

// GetErrors возвращает список ошибок
func (elp *ErrorLogPanel) GetErrors() []LogEntry {
	return elp.errors
}

// SetFilter устанавливает фильтр
func (elp *ErrorLogPanel) SetFilter(text string, level string) {
	elp.filterText = text
	elp.filterLevel = level
	elp.selectedIndex = 0
}
