package ui

// Клавиатурные управления
const (
	KeyQuit   = "q"
	KeyPause  = "p"
	KeyStop   = "s"
	KeyReset  = "r"
	KeyHelp   = "h"
	KeyTab    = "tab"
	KeyEnter  = "enter"
	KeyEscape = "esc"
	KeyCtrlC  = "ctrl+c"
	KeyUp     = "up"
	KeyDown   = "down"
	KeyLeft   = "left"
	KeyRight  = "right"
	KeySpace  = " "
)

// Ограничения производительности
const (
	MaxResults     = 1000 // Максимальное количество результатов в памяти
	MaxErrors      = 100  // Максимальное количество ошибок в логе
	MaxRPSHistory  = 60   // Количество секунд истории RPS
	UpdateInterval = 100  // Интервал обновления в миллисекундах
)

// Минимальные размеры терминала
const (
	MinWidth  = 80
	MinHeight = 24
)

// Состояния тестирования
type TestStatus string

const (
	StatusStopped  TestStatus = "Stopped"
	StatusRunning  TestStatus = "Running"
	StatusPaused   TestStatus = "Paused"
	StatusFinished TestStatus = "Finished"
	StatusError    TestStatus = "Error"
)

// HTTP методы
var HTTPMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

// Статус коды для отображения
var StatusCodes = []int{200, 201, 204, 301, 302, 400, 401, 403, 404, 429, 500, 502, 503, 504}

// Форматирование времени
const (
	TimeFormat = "15:04:05"
	DateFormat = "2006-01-02 15:04:05"
)

// Размеры компонентов
const (
	ConfigPanelHeight      = 6
	MetricsPanelHeight     = 8
	ProgressPanelHeight    = 3
	StatusCodesPanelHeight = 6
	ErrorLogPanelHeight    = 8
	ResponseChartHeight    = 8
)

// Символы для графиков
const (
	BlockChar      = "█"
	LightBlockChar = "░"
	BarChar        = "▊"
	DotChar        = "•"
	ArrowUp        = "↑"
	ArrowDown      = "↓"
	ArrowRight     = "→"
	ArrowLeft      = "←"
)

// Сообщения помощи
const HelpMessage = `
Controls:
  q - Quit application
  p - Pause/Resume test
  s - Stop test
  r - Reset test
  h - Show this help
  Tab - Switch between panels
  Ctrl+C - Force quit
`

// Ошибки терминала
const (
	ErrTerminalTooSmall = "Terminal too small. Minimum size: 80x24"
	ErrInvalidConfig    = "Invalid configuration"
	ErrTestFailed       = "Test execution failed"
)
