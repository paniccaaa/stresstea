package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/paniccaaa/stresstea/internal/loadtest"
)

// Component представляет базовый интерфейс для UI компонентов
type Component interface {
	// Render отображает компонент
	Render(width, height int) string

	// Update обновляет состояние компонента
	Update(msg tea.Msg) Component

	// SetSize устанавливает размер компонента
	SetSize(width, height int)

	// GetSize возвращает размер компонента
	GetSize() (width, height int)

	// Focus устанавливает фокус на компонент
	Focus() Component

	// Blur убирает фокус с компонента
	Blur() Component

	// IsFocused возвращает состояние фокуса
	IsFocused() bool
}

// BaseComponent содержит общие поля для всех компонентов
type BaseComponent struct {
	width   int
	height  int
	focused bool
}

// SetSize устанавливает размер компонента
func (b *BaseComponent) SetSize(width, height int) {
	b.width = width
	b.height = height
}

// GetSize возвращает размер компонента
func (b *BaseComponent) GetSize() (int, int) {
	return b.width, b.height
}

// Focus устанавливает фокус на компонент
func (b *BaseComponent) Focus() Component {
	b.focused = true
	return b
}

// Blur убирает фокус с компонента
func (b *BaseComponent) Blur() Component {
	b.focused = false
	return b
}

// Render базовый рендер (должен быть переопределен)
func (b *BaseComponent) Render(width, height int) string {
	return ""
}

// Update базовое обновление (должно быть переопределено)
func (b *BaseComponent) Update(msg tea.Msg) Component {
	return b
}

// IsFocused возвращает состояние фокуса
func (b *BaseComponent) IsFocused() bool {
	return b.focused
}

// MetricsProvider предоставляет метрики для компонентов
type MetricsProvider interface {
	GetMetrics() *Metrics
	GetResults() []loadtest.Result
	GetConfig() *Config
}

// Config содержит конфигурацию для компонентов
type Config struct {
	Target     string
	Method     string
	Duration   time.Duration
	Rate       int
	Concurrent int
	Protocol   string
}

// LayoutManager управляет расположением компонентов
type LayoutManager interface {
	// AddComponent добавляет компонент в layout
	AddComponent(name string, component Component) error

	// RemoveComponent удаляет компонент из layout
	RemoveComponent(name string) error

	// GetComponent возвращает компонент по имени
	GetComponent(name string) (Component, bool)

	// Render отображает весь layout
	Render(width, height int) string

	// Update обновляет все компоненты
	Update(msg tea.Msg) LayoutManager

	// SetFocus устанавливает фокус на компонент
	SetFocus(componentName string) error

	// GetFocusedComponent возвращает сфокусированный компонент
	GetFocusedComponent() Component
}
