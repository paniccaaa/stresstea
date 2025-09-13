package ui

import (
	"sort"
	"time"

	"github.com/paniccaaa/stresstea/internal/loadtest"
	"github.com/paniccaaa/stresstea/internal/parser"
)

// Metrics содержит расширенные метрики для нагрузочного тестирования
type Metrics struct {
	config *parser.Config

	// Основные метрики
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	SuccessRate        float64

	// Время отклика
	AvgLatency time.Duration
	MinLatency time.Duration
	MaxLatency time.Duration
	P50Latency time.Duration
	P90Latency time.Duration
	P95Latency time.Duration
	P99Latency time.Duration

	// RPS метрики
	CurrentRPS float64
	TargetRPS  int
	RPSHistory []float64 // Последние 60 секунд для графика

	// Статус коды
	StatusCodes map[int]int

	// Ошибки
	RecentErrors []string // Последние 10 ошибок

	// Время
	StartTime     time.Time
	ElapsedTime   time.Duration
	RemainingTime time.Duration

	// Throughput
	BytesPerSecond int64
	TotalBytes     int64

	// Дополнительные метрики
	RequestsPerSecond float64
	ErrorRate         float64
	ThroughputMBps    float64

	// Состояние теста
	status TestStatus
}

// NewMetrics создает новую структуру метрик
func NewMetrics(config *parser.Config) *Metrics {
	return &Metrics{
		config:       config,
		StatusCodes:  make(map[int]int),
		RPSHistory:   make([]float64, 0, MaxRPSHistory),
		RecentErrors: make([]string, 0, MaxErrors),
		TargetRPS:    config.Test.Rate,
		StartTime:    time.Now(),
	}
}

// UpdateMetrics обновляет метрики на основе новых результатов
func (m *Metrics) UpdateMetrics(results []loadtest.Result) {
	if len(results) == 0 {
		return
	}

	// Ограничиваем количество результатов в памяти
	if len(results) > MaxResults {
		results = results[len(results)-MaxResults:]
	}

	// Сбрасываем метрики
	m.TotalRequests = len(results)
	m.SuccessfulRequests = 0
	m.FailedRequests = 0
	m.StatusCodes = make(map[int]int)

	var latencies []time.Duration
	var totalLatency time.Duration
	var totalBytes int64

	// Обрабатываем результаты
	for _, result := range results {
		if result.Error != nil {
			m.FailedRequests++
			// Добавляем ошибку в лог (максимум MaxErrors)
			if len(m.RecentErrors) >= MaxErrors {
				m.RecentErrors = m.RecentErrors[1:]
			}
			m.RecentErrors = append(m.RecentErrors, result.Error.Error())
		} else {
			m.SuccessfulRequests++
			latencies = append(latencies, result.Latency)
			totalLatency += result.Latency
		}

		// Статус коды
		if result.Status > 0 {
			m.StatusCodes[result.Status]++
		}

		// Байты
		totalBytes += result.Bytes
	}

	// Успешность
	if m.TotalRequests > 0 {
		m.SuccessRate = float64(m.SuccessfulRequests) / float64(m.TotalRequests) * 100
		m.ErrorRate = float64(m.FailedRequests) / float64(m.TotalRequests) * 100
	}

	// Время отклика
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		m.AvgLatency = totalLatency / time.Duration(len(latencies))
		m.MinLatency = latencies[0]
		m.MaxLatency = latencies[len(latencies)-1]

		// Percentiles
		m.P50Latency = m.calculatePercentile(latencies, 50)
		m.P90Latency = m.calculatePercentile(latencies, 90)
		m.P95Latency = m.calculatePercentile(latencies, 95)
		m.P99Latency = m.calculatePercentile(latencies, 99)
	}

	// Время
	m.ElapsedTime = time.Since(m.StartTime)
	if m.config != nil {
		m.RemainingTime = m.config.Test.Duration - m.ElapsedTime
		if m.RemainingTime < 0 {
			m.RemainingTime = 0
		}
	}

	// RPS
	if m.ElapsedTime.Seconds() > 0 {
		m.CurrentRPS = float64(m.TotalRequests) / m.ElapsedTime.Seconds()
		m.RequestsPerSecond = m.CurrentRPS
	}

	// Throughput
	if m.ElapsedTime.Seconds() > 0 {
		m.BytesPerSecond = int64(float64(totalBytes) / m.ElapsedTime.Seconds())
		m.ThroughputMBps = float64(m.BytesPerSecond) / (1024 * 1024)
	}
	m.TotalBytes = totalBytes

	// Обновляем историю RPS
	m.updateRPSHistory()
}

// calculatePercentile вычисляет перцентиль для массива latencies
func (m *Metrics) calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	index := int(float64(len(latencies)) * float64(percentile) / 100.0)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return latencies[index]
}

// updateRPSHistory обновляет историю RPS
func (m *Metrics) updateRPSHistory() {
	// Добавляем текущий RPS в историю
	m.RPSHistory = append(m.RPSHistory, m.CurrentRPS)

	// Ограничиваем размер истории
	if len(m.RPSHistory) > MaxRPSHistory {
		m.RPSHistory = m.RPSHistory[1:]
	}
}

// GetProgress возвращает прогресс выполнения теста (0.0 - 1.0)
func (m *Metrics) GetProgress() float64 {
	if m.config == nil || m.config.Test.Duration == 0 {
		return 0.0
	}

	progress := float64(m.ElapsedTime) / float64(m.config.Test.Duration)
	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}

// IsTestFinished возвращает true, если тест завершен
func (m *Metrics) IsTestFinished() bool {
	return m.RemainingTime <= 0
}

// GetStatusCodesSorted возвращает статус коды, отсортированные по количеству
func (m *Metrics) GetStatusCodesSorted() []StatusCodeInfo {
	var result []StatusCodeInfo

	for status, count := range m.StatusCodes {
		percentage := 0.0
		if m.TotalRequests > 0 {
			percentage = float64(count) / float64(m.TotalRequests) * 100
		}

		result = append(result, StatusCodeInfo{
			Status:     status,
			Count:      count,
			Percentage: percentage,
		})
	}

	// Сортируем по количеству (убывание)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

// StatusCodeInfo содержит информацию о статус коде
type StatusCodeInfo struct {
	Status     int
	Count      int
	Percentage float64
}

// GetTopErrors возвращает топ ошибок
func (m *Metrics) GetTopErrors() []string {
	if len(m.RecentErrors) == 0 {
		return []string{"No errors"}
	}

	// Возвращаем последние ошибки
	start := 0
	if len(m.RecentErrors) > 10 {
		start = len(m.RecentErrors) - 10
	}

	return m.RecentErrors[start:]
}

// SetStatus устанавливает статус теста
func (m *Metrics) SetStatus(status TestStatus) {
	m.status = status
}

// GetStatus возвращает статус теста
func (m *Metrics) GetStatus() TestStatus {
	return m.status
}
