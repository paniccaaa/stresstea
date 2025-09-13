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

	// Скользящее окно для точного RPS
	requestTimestamps []time.Time
	windowSize        time.Duration // Размер окна для расчета RPS (по умолчанию 1 секунда)

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
		config:            config,
		StatusCodes:       make(map[int]int),
		RPSHistory:        make([]float64, 0, MaxRPSHistory),
		RecentErrors:      make([]string, 0, MaxErrors),
		TargetRPS:         config.Test.Rate,
		StartTime:         time.Now(),
		requestTimestamps: make([]time.Time, 0, 1000),
		windowSize:        time.Second, // 1 секунда для расчета RPS
	}
}

// UpdateMetrics обновляет метрики на основе новых результатов
func (m *Metrics) UpdateMetrics(results []loadtest.Result) {
	if len(results) == 0 {
		return
	}

	// Валидация входных данных
	if m.config == nil {
		return
	}

	// Ограничиваем количество результатов в памяти
	if len(results) > MaxResults {
		results = results[len(results)-MaxResults:]
	}

	// НЕ сбрасываем метрики, а обновляем их
	m.TotalRequests += len(results)

	var latencies []time.Duration
	var totalLatency time.Duration
	var totalBytes int64

	// Обрабатываем результаты
	for _, result := range results {
		// Добавляем timestamp для расчета RPS
		m.requestTimestamps = append(m.requestTimestamps, result.Timestamp)

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

	// Успешность с валидацией
	if m.TotalRequests > 0 {
		m.SuccessRate = float64(m.SuccessfulRequests) / float64(m.TotalRequests) * 100
		m.ErrorRate = float64(m.FailedRequests) / float64(m.TotalRequests) * 100

		// Валидация: сумма успешных и неудачных не должна превышать общее количество
		if m.SuccessfulRequests+m.FailedRequests > m.TotalRequests {
			// Корректируем если есть несоответствие
			m.TotalRequests = m.SuccessfulRequests + m.FailedRequests
		}
	}

	// Время отклика - обновляем только если есть новые успешные запросы
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		// Обновляем метрики латентности
		m.AvgLatency = totalLatency / time.Duration(len(latencies))
		if m.MinLatency == 0 || latencies[0] < m.MinLatency {
			m.MinLatency = latencies[0]
		}
		if latencies[len(latencies)-1] > m.MaxLatency {
			m.MaxLatency = latencies[len(latencies)-1]
		}

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

	// RPS - используем скользящее окно
	m.CurrentRPS = m.calculateCurrentRPS()
	m.RequestsPerSecond = m.CurrentRPS

	// Throughput
	m.TotalBytes += totalBytes
	if m.ElapsedTime.Seconds() > 0 {
		m.BytesPerSecond = int64(float64(m.TotalBytes) / m.ElapsedTime.Seconds())
		m.ThroughputMBps = float64(m.BytesPerSecond) / (1024 * 1024)
	}

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

// calculateCurrentRPS вычисляет текущий RPS используя скользящее окно
func (m *Metrics) calculateCurrentRPS() float64 {
	if len(m.requestTimestamps) == 0 {
		return 0
	}

	now := time.Now()
	windowStart := now.Add(-m.windowSize)

	// Подсчитываем запросы в окне
	count := 0
	for _, timestamp := range m.requestTimestamps {
		if timestamp.After(windowStart) {
			count++
		}
	}

	// Очищаем старые timestamps
	m.cleanOldTimestamps(now)

	// Возвращаем RPS для окна
	return float64(count) / m.windowSize.Seconds()
}

// cleanOldTimestamps удаляет timestamps старше окна
func (m *Metrics) cleanOldTimestamps(now time.Time) {
	windowStart := now.Add(-m.windowSize)

	// Находим первый индекс, который нужно сохранить
	keepFrom := 0
	for i, timestamp := range m.requestTimestamps {
		if timestamp.After(windowStart) {
			keepFrom = i
			break
		}
	}

	// Удаляем старые timestamps
	if keepFrom > 0 {
		m.requestTimestamps = m.requestTimestamps[keepFrom:]
	}
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
