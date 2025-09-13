package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/paniccaaa/stresstea/internal/parser"
)

// ConfigPanel отображает конфигурацию теста с интерактивными элементами
type ConfigPanel struct {
	BaseComponent
	config *parser.Config

	// Компоненты ввода
	urlInput      textinput.Model
	methodList    list.Model
	durationInput textinput.Model
	rateInput     textinput.Model
	threadsInput  textinput.Model

	// Состояние
	focusedField string
	editing      bool
}

// NewConfigPanel создает новую панель конфигурации
func NewConfigPanel(config *parser.Config) *ConfigPanel {
	cp := &ConfigPanel{
		BaseComponent: BaseComponent{},
		config:        config,
		focusedField:  "url",
		editing:       false,
	}

	cp.initComponents()
	return cp
}

// initComponents инициализирует компоненты ввода
func (cp *ConfigPanel) initComponents() {
	// URL input
	cp.urlInput = textinput.New()
	cp.urlInput.Placeholder = "Enter target URL..."
	cp.urlInput.CharLimit = 200
	cp.urlInput.Width = 40
	cp.urlInput.SetValue(cp.config.Test.Target)
	cp.urlInput.Focus()

	// Method list
	cp.methodList = list.New(createMethodItems(), list.NewDefaultDelegate(), 0, 0)
	cp.methodList.Title = "HTTP Method"
	cp.methodList.SetShowStatusBar(false)
	cp.methodList.SetFilteringEnabled(false)
	cp.methodList.DisableQuitKeybindings()

	// Set selected method
	for i, item := range cp.methodList.Items() {
		if item.(methodItem).title == cp.config.Test.Method {
			cp.methodList.Select(i)
			break
		}
	}

	// Duration input
	cp.durationInput = textinput.New()
	cp.durationInput.Placeholder = "e.g., 30s, 1m, 2h..."
	cp.durationInput.CharLimit = 10
	cp.durationInput.Width = 15
	cp.durationInput.SetValue(cp.config.Test.Duration.String())

	// Rate input
	cp.rateInput = textinput.New()
	cp.rateInput.Placeholder = "Requests per second"
	cp.rateInput.CharLimit = 10
	cp.rateInput.Width = 15
	cp.rateInput.SetValue(strconv.Itoa(cp.config.Test.Rate))

	// Threads input
	cp.threadsInput = textinput.New()
	cp.threadsInput.Placeholder = "Concurrent threads"
	cp.threadsInput.CharLimit = 10
	cp.threadsInput.Width = 15
	cp.threadsInput.SetValue(strconv.Itoa(cp.config.Test.Concurrent))
}

// Render отображает панель конфигурации
func (cp *ConfigPanel) Render(width, height int) string {
	cp.SetSize(width, height)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1).
		Render("Configuration")

	// URL section
	urlLabel := lipgloss.NewStyle().Bold(true).Render("Target URL:")
	urlInput := cp.renderInput(cp.urlInput, "url")

	// Method section
	methodLabel := lipgloss.NewStyle().Bold(true).Render("HTTP Method:")
	methodList := cp.renderMethodList()

	// Parameters section
	paramsLabel := lipgloss.NewStyle().Bold(true).Render("Parameters:")
	durationLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("Duration:")
	durationInput := cp.renderInput(cp.durationInput, "duration")

	rateLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("Rate (RPS):")
	rateInput := cp.renderInput(cp.rateInput, "rate")

	threadsLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("Threads:")
	threadsInput := cp.renderInput(cp.threadsInput, "threads")

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Render("Tab: Next field | Enter: Edit | Esc: Save")

	content := fmt.Sprintf(`%s

%s
%s

%s
%s

%s
%s %s
%s %s
%s %s

%s`,
		title,
		urlLabel, urlInput,
		methodLabel, methodList,
		paramsLabel,
		durationLabel, durationInput,
		rateLabel, rateInput,
		threadsLabel, threadsInput,
		helpText)

	return PanelStyle.Render(content)
}

// renderInput отображает поле ввода с фокусом
func (cp *ConfigPanel) renderInput(input textinput.Model, fieldName string) string {
	style := lipgloss.NewStyle().Padding(0, 1)

	if cp.focusedField == fieldName && cp.editing {
		style = style.Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))
	} else if cp.focusedField == fieldName {
		style = style.Background(lipgloss.Color("#3C3C3C"))
	}

	return style.Render(input.View())
}

// renderMethodList отображает список методов
func (cp *ConfigPanel) renderMethodList() string {
	if cp.focusedField == "method" {
		return cp.methodList.View()
	}

	// Показываем только выбранный метод
	selected := cp.methodList.SelectedItem()
	if selected != nil {
		return lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("#3C3C3C")).
			Render(selected.(methodItem).title)
	}

	return lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("#626262")).
		Render("No method selected")
}

// Update обновляет панель конфигурации
func (cp *ConfigPanel) Update(msg tea.Msg) Component {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case KeyTab:
			cp.switchField()
		case KeyEnter:
			if !cp.editing {
				cp.editing = true
				cp.focusCurrentField()
			} else {
				cp.editing = false
				cp.blurAllFields()
				cp.updateConfig()
			}
		case KeyEscape:
			if cp.editing {
				cp.editing = false
				cp.blurAllFields()
			}
		}
	}

	// Обновляем активное поле
	if cp.editing {
		switch cp.focusedField {
		case "url":
			cp.urlInput, _ = cp.urlInput.Update(msg)
		case "duration":
			cp.durationInput, _ = cp.durationInput.Update(msg)
		case "rate":
			cp.rateInput, _ = cp.rateInput.Update(msg)
		case "threads":
			cp.threadsInput, _ = cp.threadsInput.Update(msg)
		case "method":
			cp.methodList, _ = cp.methodList.Update(msg)
		}
	}

	return cp
}

// switchField переключает между полями
func (cp *ConfigPanel) switchField() {
	fields := []string{"url", "method", "duration", "rate", "threads"}
	currentIndex := 0

	for i, field := range fields {
		if field == cp.focusedField {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(fields)
	cp.focusedField = fields[nextIndex]
}

// focusCurrentField устанавливает фокус на текущее поле
func (cp *ConfigPanel) focusCurrentField() {
	cp.blurAllFields()

	switch cp.focusedField {
	case "url":
		cp.urlInput.Focus()
	case "duration":
		cp.durationInput.Focus()
	case "rate":
		cp.rateInput.Focus()
	case "threads":
		cp.threadsInput.Focus()
	case "method":
		// Method list doesn't need focus
	}
}

// blurAllFields убирает фокус со всех полей
func (cp *ConfigPanel) blurAllFields() {
	cp.urlInput.Blur()
	cp.durationInput.Blur()
	cp.rateInput.Blur()
	cp.threadsInput.Blur()
}

// updateConfig обновляет конфигурацию из полей ввода
func (cp *ConfigPanel) updateConfig() {
	// URL
	cp.config.Test.Target = cp.urlInput.Value()

	// Method
	if selected := cp.methodList.SelectedItem(); selected != nil {
		cp.config.Test.Method = selected.(methodItem).title
	}

	// Duration
	if duration, err := time.ParseDuration(cp.durationInput.Value()); err == nil {
		cp.config.Test.Duration = duration
	}

	// Rate
	if rate, err := strconv.Atoi(cp.rateInput.Value()); err == nil && rate > 0 {
		cp.config.Test.Rate = rate
	}

	// Threads
	if threads, err := strconv.Atoi(cp.threadsInput.Value()); err == nil && threads > 0 {
		cp.config.Test.Concurrent = threads
	}
}

// methodItem представляет элемент списка методов
type methodItem struct {
	title, desc string
}

func (i methodItem) Title() string       { return i.title }
func (i methodItem) Description() string { return i.desc }
func (i methodItem) FilterValue() string { return i.title }

// createMethodItems создает элементы списка HTTP методов
func createMethodItems() []list.Item {
	items := []list.Item{
		methodItem{title: "GET", desc: "Retrieve data from server"},
		methodItem{title: "POST", desc: "Send data to server"},
		methodItem{title: "PUT", desc: "Update existing resource"},
		methodItem{title: "DELETE", desc: "Delete resource"},
		methodItem{title: "PATCH", desc: "Partially update resource"},
		methodItem{title: "HEAD", desc: "Get headers only"},
		methodItem{title: "OPTIONS", desc: "Get allowed methods"},
	}

	return items
}

// GetConfig возвращает текущую конфигурацию
func (cp *ConfigPanel) GetConfig() *parser.Config {
	cp.updateConfig()
	return cp.config
}
