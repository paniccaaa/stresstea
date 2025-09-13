package ui

import "github.com/charmbracelet/lipgloss"

// Цветовая схема для TUI интерфейса
var (
	// Основные цвета
	Primary = lipgloss.Color("#7D56F4") // Фиолетовый
	Success = lipgloss.Color("#00FF00") // Зеленый
	Error   = lipgloss.Color("#FF0000") // Красный
	Warning = lipgloss.Color("#FFA500") // Оранжевый
	Info    = lipgloss.Color("#00BFFF") // Голубой
	Muted   = lipgloss.Color("#626262") // Серый
	White   = lipgloss.Color("#FAFAFA") // Белый
	Black   = lipgloss.Color("#000000") // Черный

	// Градиенты
	PrimaryGradient = []string{"#7D56F4", "#874BFD", "#9D6BFF"}
	SuccessGradient = []string{"#00FF00", "#32CD32", "#00CC00"}
	ErrorGradient   = []string{"#FF0000", "#FF4500", "#FF6347"}
)

// Стили для различных элементов
var (
	// Заголовок
	TitleStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Primary).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	// Панели
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2).
			Margin(0, 1, 0, 0)

	// Успешные элементы
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	// Ошибки
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Предупреждения
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	// Информация
	InfoStyle = lipgloss.NewStyle().
			Foreground(Info).
			Bold(true)

	// Приглушенный текст
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Помощь
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Прогресс бар
	ProgressStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Background(Muted)

	// Таблица
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(White).
				Background(Primary).
				Bold(true).
				Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Статус коды
	Status200Style = lipgloss.NewStyle().Foreground(Success)
	Status300Style = lipgloss.NewStyle().Foreground(Info)
	Status400Style = lipgloss.NewStyle().Foreground(Warning)
	Status500Style = lipgloss.NewStyle().Foreground(Error)
)

// GetStatusStyle возвращает стиль для статус кода
func GetStatusStyle(status int) lipgloss.Style {
	switch {
	case status >= 200 && status < 300:
		return Status200Style
	case status >= 300 && status < 400:
		return Status300Style
	case status >= 400 && status < 500:
		return Status400Style
	case status >= 500:
		return Status500Style
	default:
		return MutedStyle
	}
}
