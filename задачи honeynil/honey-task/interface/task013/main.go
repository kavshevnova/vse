package main

// Задача: Metrics — счётчики, gauge, гистограммы.

import (
	"fmt"
	"sync"
	"time"
)

type MetricType string

const (
	Counter   MetricType = "counter"
	Gauge     MetricType = "gauge"
	Histogram MetricType = "histogram"
)

type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

type Metrics interface {
	Inc(name string, labels map[string]string)
	Add(name string, value float64, labels map[string]string)
	Set(name string, value float64, labels map[string]string)
	Observe(name string, value float64, labels map[string]string)
	GetAll() []Metric
	Reset()
}

type SimpleMetrics struct {
	mu      sync.Mutex
	metrics []Metric
}

func (m *SimpleMetrics) record(name string, t MetricType, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, Metric{
		Name: name, Type: t, Value: value, Labels: labels, Timestamp: time.Now(),
	})
}

func (m *SimpleMetrics) Inc(name string, labels map[string]string)                    { m.Add(name, 1, labels) }
func (m *SimpleMetrics) Add(name string, value float64, labels map[string]string)     { m.record(name, Counter, value, labels) }
func (m *SimpleMetrics) Set(name string, value float64, labels map[string]string)     { m.record(name, Gauge, value, labels) }
func (m *SimpleMetrics) Observe(name string, value float64, labels map[string]string) { m.record(name, Histogram, value, labels) }

func (m *SimpleMetrics) GetAll() []Metric {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Metric, len(m.metrics))
	copy(out, m.metrics)
	return out
}

func (m *SimpleMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = nil
}

// --- BufferedMetrics ---

type BufferedMetrics struct {
	underlying    Metrics
	buf           []Metric
	bufferSize    int
	flushInterval time.Duration
	mu            sync.Mutex
	ticker        *time.Ticker
	done          chan struct{}
}

func NewBufferedMetrics(underlying Metrics, bufferSize int, flushInterval time.Duration) *BufferedMetrics {
	bm := &BufferedMetrics{
		underlying:    underlying,
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		done:          make(chan struct{}),
	}
	bm.ticker = time.NewTicker(flushInterval)
	go bm.flushLoop()
	return bm
}

func (b *BufferedMetrics) add(m Metric) {
	b.mu.Lock()
	b.buf = append(b.buf, m)
	flush := len(b.buf) >= b.bufferSize
	b.mu.Unlock()
	if flush {
		b.flush()
	}
}

func (b *BufferedMetrics) flush() {
	b.mu.Lock()
	buf := b.buf
	b.buf = nil
	b.mu.Unlock()
	for _, m := range buf {
		switch m.Type {
		case Counter:
			b.underlying.Add(m.Name, m.Value, m.Labels)
		case Gauge:
			b.underlying.Set(m.Name, m.Value, m.Labels)
		case Histogram:
			b.underlying.Observe(m.Name, m.Value, m.Labels)
		}
	}
}

func (b *BufferedMetrics) flushLoop() {
	for {
		select {
		case <-b.ticker.C:
			b.flush()
		case <-b.done:
			b.flush()
			return
		}
	}
}

func (b *BufferedMetrics) Inc(name string, labels map[string]string) {
	b.add(Metric{Name: name, Type: Counter, Value: 1, Labels: labels, Timestamp: time.Now()})
}
func (b *BufferedMetrics) Add(name string, v float64, labels map[string]string) {
	b.add(Metric{Name: name, Type: Counter, Value: v, Labels: labels, Timestamp: time.Now()})
}
func (b *BufferedMetrics) Set(name string, v float64, labels map[string]string) {
	b.add(Metric{Name: name, Type: Gauge, Value: v, Labels: labels, Timestamp: time.Now()})
}
func (b *BufferedMetrics) Observe(name string, v float64, labels map[string]string) {
	b.add(Metric{Name: name, Type: Histogram, Value: v, Labels: labels, Timestamp: time.Now()})
}
func (b *BufferedMetrics) GetAll() []Metric  { return b.underlying.GetAll() }
func (b *BufferedMetrics) Reset()            { b.underlying.Reset() }
func (b *BufferedMetrics) Close() {
	b.ticker.Stop()
	close(b.done)
}

func main() {
	m := &SimpleMetrics{}
	m.Inc("requests", map[string]string{"method": "GET"})
	m.Set("memory_bytes", 1024*1024, nil)
	m.Observe("latency_ms", 42.5, map[string]string{"endpoint": "/api"})
	for _, metric := range m.GetAll() {
		fmt.Printf("%s(%s)=%.1f\n", metric.Name, metric.Type, metric.Value)
	}
}
