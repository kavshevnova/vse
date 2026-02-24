package main

// Задача: Streaming Pipeline — backpressure, windowing, watermarks, checkpointing.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Message struct {
	Key       string
	Value     interface{}
	Timestamp time.Time
	Metadata  map[string]string
}

type WindowType string

const (
	TumblingWindow WindowType = "tumbling"
	SlidingWindow  WindowType = "sliding"
	SessionWindow  WindowType = "session"
)

type Window struct {
	Start time.Time
	End   time.Time
	Type  WindowType
}

type Processor interface {
	Process(ctx context.Context, msg Message) ([]Message, error)
}

type Aggregator interface {
	Aggregate(ctx context.Context, window Window, messages []Message) (interface{}, error)
}

type StreamSource interface {
	Read(ctx context.Context, maxBatch int) ([]Message, error)
	Commit(ctx context.Context, offset int64) error
	Seek(ctx context.Context, offset int64) error
}

type StreamSink interface {
	Write(ctx context.Context, messages []Message) error
	Flush(ctx context.Context) error
}

type Pipeline interface {
	AddSource(name string, source StreamSource) error
	AddProcessor(name string, processor Processor) error
	AddSink(name string, sink StreamSink) error
	AddWindow(name string, windowType WindowType, size time.Duration, aggregator Aggregator) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Checkpoint(ctx context.Context) error
	Restore(ctx context.Context, checkpointID string) error
}

type BackpressureStrategy interface {
	ShouldBlock(ctx context.Context, queueSize int, maxSize int) bool
	OnBackpressure(ctx context.Context) error
}

type Watermark interface {
	UpdateWatermark(timestamp time.Time)
	GetWatermark() time.Time
	IsLate(msgTimestamp time.Time) bool
}

// --- InMemoryStreamSource ---

type InMemoryStreamSource struct {
	messages []Message
	offset   int64
	mu       sync.Mutex
}

func NewInMemorySource(msgs []Message) *InMemoryStreamSource {
	return &InMemoryStreamSource{messages: msgs}
}

func (s *InMemoryStreamSource) Read(_ context.Context, maxBatch int) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	start := int(s.offset)
	if start >= len(s.messages) {
		return nil, nil
	}
	end := start + maxBatch
	if end > len(s.messages) {
		end = len(s.messages)
	}
	batch := s.messages[start:end]
	return batch, nil
}

func (s *InMemoryStreamSource) Commit(_ context.Context, offset int64) error {
	s.mu.Lock()
	s.offset = offset
	s.mu.Unlock()
	return nil
}

func (s *InMemoryStreamSource) Seek(_ context.Context, offset int64) error {
	s.mu.Lock()
	s.offset = offset
	s.mu.Unlock()
	return nil
}

// --- InMemoryStreamSink ---

type InMemoryStreamSink struct {
	mu       sync.Mutex
	received []Message
}

func (s *InMemoryStreamSink) Write(_ context.Context, msgs []Message) error {
	s.mu.Lock()
	s.received = append(s.received, msgs...)
	s.mu.Unlock()
	return nil
}
func (s *InMemoryStreamSink) Flush(_ context.Context) error { return nil }
func (s *InMemoryStreamSink) Messages() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Message{}, s.received...)
}

// --- WatermarkImpl ---

type WatermarkImpl struct {
	mu        sync.RWMutex
	watermark time.Time
	maxLate   time.Duration
}

func NewWatermark(maxLate time.Duration) *WatermarkImpl { return &WatermarkImpl{maxLate: maxLate} }

func (w *WatermarkImpl) UpdateWatermark(ts time.Time) {
	w.mu.Lock()
	if ts.After(w.watermark) {
		w.watermark = ts
	}
	w.mu.Unlock()
}
func (w *WatermarkImpl) GetWatermark() time.Time {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.watermark
}
func (w *WatermarkImpl) IsLate(ts time.Time) bool {
	wm := w.GetWatermark()
	return ts.Before(wm.Add(-w.maxLate))
}

// --- BackpressureController ---

type DropBackpressure struct{}

func (b *DropBackpressure) ShouldBlock(_ context.Context, queueSize, maxSize int) bool {
	return queueSize >= maxSize
}
func (b *DropBackpressure) OnBackpressure(_ context.Context) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

// --- WindowManager ---

type windowDef struct {
	name       string
	windowType WindowType
	size       time.Duration
	agg        Aggregator
	buf        []Message
	lastFlush  time.Time
}

// --- StreamingPipeline ---

type pipelineCheckpoint struct {
	id      string
	offsets map[string]int64
}

type StreamingPipeline struct {
	mu          sync.Mutex
	sources     map[string]StreamSource
	processors  []Processor
	sinks       map[string]StreamSink
	windows     []*windowDef
	bp          BackpressureStrategy
	watermark   Watermark
	maxQueue    int
	checkpoints []pipelineCheckpoint
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewStreamingPipeline(maxQueue int) *StreamingPipeline {
	return &StreamingPipeline{
		sources:    make(map[string]StreamSource),
		sinks:      make(map[string]StreamSink),
		bp:         &DropBackpressure{},
		watermark:  NewWatermark(5 * time.Second),
		maxQueue:   maxQueue,
	}
}

func (p *StreamingPipeline) AddSource(name string, src StreamSource) error {
	p.mu.Lock()
	p.sources[name] = src
	p.mu.Unlock()
	return nil
}

func (p *StreamingPipeline) AddProcessor(_ string, proc Processor) error {
	p.mu.Lock()
	p.processors = append(p.processors, proc)
	p.mu.Unlock()
	return nil
}

func (p *StreamingPipeline) AddSink(name string, sink StreamSink) error {
	p.mu.Lock()
	p.sinks[name] = sink
	p.mu.Unlock()
	return nil
}

func (p *StreamingPipeline) AddWindow(name string, wt WindowType, size time.Duration, agg Aggregator) error {
	p.mu.Lock()
	p.windows = append(p.windows, &windowDef{
		name: name, windowType: wt, size: size, agg: agg, lastFlush: time.Now(),
	})
	p.mu.Unlock()
	return nil
}

func (p *StreamingPipeline) process(ctx context.Context, msg Message) []Message {
	msgs := []Message{msg}
	for _, proc := range p.processors {
		var next []Message
		for _, m := range msgs {
			out, err := proc.Process(ctx, m)
			if err == nil {
				next = append(next, out...)
			}
		}
		msgs = next
	}
	return msgs
}

func (p *StreamingPipeline) flushWindows(ctx context.Context) {
	now := time.Now()
	for _, w := range p.windows {
		if now.Sub(w.lastFlush) < w.size {
			continue
		}
		win := Window{Start: w.lastFlush, End: now, Type: w.windowType}
		result, err := w.agg.Aggregate(ctx, win, w.buf)
		if err == nil && result != nil {
			outMsg := Message{Key: w.name, Value: result, Timestamp: now}
			for _, sink := range p.sinks {
				sink.Write(ctx, []Message{outMsg})
			}
		}
		w.buf = nil
		w.lastFlush = now
	}
}

func (p *StreamingPipeline) Start(ctx context.Context) error {
	ctx, p.cancel = context.WithCancel(ctx)
	for name, src := range p.sources {
		name, src := name, src
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			offset := int64(0)
			for {
				if ctx.Err() != nil {
					return
				}
				msgs, err := src.Read(ctx, 100)
				if err != nil || len(msgs) == 0 {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				_ = name
				for _, msg := range msgs {
					if p.watermark.IsLate(msg.Timestamp) {
						continue
					}
					p.watermark.UpdateWatermark(msg.Timestamp)
					processed := p.process(ctx, msg)
					for _, sink := range p.sinks {
						sink.Write(ctx, processed)
					}
					for _, w := range p.windows {
						w.buf = append(w.buf, msg)
					}
					offset++
				}
				src.Commit(ctx, offset)
				p.flushWindows(ctx)
			}
		}()
	}
	return nil
}

func (p *StreamingPipeline) Stop(_ context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	for _, sink := range p.sinks {
		sink.Flush(context.Background())
	}
	return nil
}

func (p *StreamingPipeline) Checkpoint(_ context.Context) error {
	cp := pipelineCheckpoint{
		id:      fmt.Sprintf("cp-%d", time.Now().UnixNano()),
		offsets: make(map[string]int64),
	}
	p.mu.Lock()
	p.checkpoints = append(p.checkpoints, cp)
	p.mu.Unlock()
	return nil
}

func (p *StreamingPipeline) Restore(_ context.Context, checkpointID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, cp := range p.checkpoints {
		if cp.id == checkpointID {
			for name, offset := range cp.offsets {
				if src, ok := p.sources[name]; ok {
					src.Seek(context.Background(), offset)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("checkpoint %q not found", checkpointID)
}

// --- Demo ---

type UpperCaseProcessor struct{}

func (u *UpperCaseProcessor) Process(_ context.Context, msg Message) ([]Message, error) {
	if s, ok := msg.Value.(string); ok {
		msg.Value = fmt.Sprintf("[%s]", s)
	}
	return []Message{msg}, nil
}

type CountAggregator struct{}

func (c *CountAggregator) Aggregate(_ context.Context, w Window, msgs []Message) (interface{}, error) {
	return fmt.Sprintf("window[%v-%v] count=%d", w.Start.Format("15:04:05"), w.End.Format("15:04:05"), len(msgs)), nil
}

func main() {
	msgs := []Message{
		{Key: "a", Value: "hello", Timestamp: time.Now()},
		{Key: "b", Value: "world", Timestamp: time.Now().Add(-time.Second)},
	}

	src := NewInMemorySource(msgs)
	sink := &InMemoryStreamSink{}

	pipeline := NewStreamingPipeline(1000)
	pipeline.AddSource("input", src)
	pipeline.AddProcessor("upper", &UpperCaseProcessor{})
	pipeline.AddSink("output", sink)
	pipeline.AddWindow("1s", TumblingWindow, 100*time.Millisecond, &CountAggregator{})

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	pipeline.Start(ctx)
	<-ctx.Done()
	pipeline.Stop(context.Background())

	for _, m := range sink.Messages() {
		fmt.Printf("sink: key=%s value=%v\n", m.Key, m.Value)
	}
}
