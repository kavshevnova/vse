package main

// Задача: Distributed Task Queue — приоритеты, retry, DLQ, exactly-once семантика.

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

type Task struct {
	ID          string
	Payload     interface{}
	Priority    Priority
	MaxRetries  int
	Timeout     time.Duration
	ScheduledAt time.Time
}

type TaskResult struct {
	TaskID    string
	Success   bool
	Result    interface{}
	Error     error
	Attempts  int
	StartedAt time.Time
	EndedAt   time.Time
}

type TaskHandler func(ctx context.Context, task Task) (interface{}, error)

type TaskQueue interface {
	Enqueue(ctx context.Context, task Task) error
	EnqueueBatch(ctx context.Context, tasks []Task) error
	Dequeue(ctx context.Context, timeout time.Duration) (*Task, error)
	Complete(ctx context.Context, taskID string, result interface{}) error
	Fail(ctx context.Context, taskID string, err error) error
	GetDeadLetterQueue(ctx context.Context, limit int) ([]Task, error)
	Stats(ctx context.Context) (QueueStats, error)
}

type QueueStats struct {
	PendingTasks    int
	ProcessingTasks int
	CompletedTasks  int
	FailedTasks     int
	DLQTasks        int
}

type Worker interface {
	RegisterHandler(taskType string, handler TaskHandler) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// --- Priority queue ---

type taskItem struct {
	task     Task
	attempts int
	index    int
}

type taskHeap []*taskItem

func (h taskHeap) Len() int { return len(h) }
func (h taskHeap) Less(i, j int) bool {
	if h[i].task.Priority != h[j].task.Priority {
		return h[i].task.Priority > h[j].task.Priority
	}
	return h[i].task.ScheduledAt.Before(h[j].task.ScheduledAt)
}
func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *taskHeap) Push(x interface{}) {
	item := x.(*taskItem)
	item.index = len(*h)
	*h = append(*h, item)
}
func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

// --- DistributedTaskQueue (in-memory) ---

type DistributedTaskQueue struct {
	mu         sync.Mutex
	pending    taskHeap
	processing map[string]*taskItem
	dlq        []Task
	results    map[string]TaskResult
	seen       map[string]bool // exactly-once
	notify     chan struct{}
}

func NewDistributedTaskQueue() *DistributedTaskQueue {
	q := &DistributedTaskQueue{
		processing: make(map[string]*taskItem),
		results:    make(map[string]TaskResult),
		seen:       make(map[string]bool),
		notify:     make(chan struct{}, 1),
	}
	heap.Init(&q.pending)
	return q
}

func (q *DistributedTaskQueue) enqueue(task Task) error {
	if q.seen[task.ID] {
		return nil // exactly-once: уже в очереди или обработана
	}
	q.seen[task.ID] = true
	if task.ScheduledAt.IsZero() {
		task.ScheduledAt = time.Now()
	}
	heap.Push(&q.pending, &taskItem{task: task})
	select {
	case q.notify <- struct{}{}:
	default:
	}
	return nil
}

func (q *DistributedTaskQueue) Enqueue(_ context.Context, task Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.enqueue(task)
}

func (q *DistributedTaskQueue) EnqueueBatch(_ context.Context, tasks []Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range tasks {
		if err := q.enqueue(t); err != nil {
			return err
		}
	}
	return nil
}

func (q *DistributedTaskQueue) Dequeue(ctx context.Context, timeout time.Duration) (*Task, error) {
	deadline := time.Now().Add(timeout)
	for {
		q.mu.Lock()
		for q.pending.Len() > 0 {
			item := heap.Pop(&q.pending).(*taskItem)
			if item.task.ScheduledAt.After(time.Now()) {
				heap.Push(&q.pending, item)
				break
			}
			q.processing[item.task.ID] = item
			q.mu.Unlock()
			return &item.task, nil
		}
		q.mu.Unlock()

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, context.DeadlineExceeded
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-q.notify:
		case <-time.After(remaining):
			return nil, context.DeadlineExceeded
		}
	}
}

func (q *DistributedTaskQueue) Complete(_ context.Context, taskID string, result interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	item, ok := q.processing[taskID]
	if !ok {
		return fmt.Errorf("task %q not in processing", taskID)
	}
	delete(q.processing, taskID)
	q.results[taskID] = TaskResult{TaskID: taskID, Success: true, Result: result, Attempts: item.attempts + 1}
	return nil
}

func (q *DistributedTaskQueue) Fail(_ context.Context, taskID string, err error) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	item, ok := q.processing[taskID]
	if !ok {
		return fmt.Errorf("task %q not in processing", taskID)
	}
	delete(q.processing, taskID)
	item.attempts++
	if item.attempts >= item.task.MaxRetries {
		q.dlq = append(q.dlq, item.task)
		q.results[taskID] = TaskResult{TaskID: taskID, Success: false, Error: err, Attempts: item.attempts}
		return nil
	}
	// Exponential backoff before re-queue
	delay := time.Duration(1<<item.attempts) * 100 * time.Millisecond
	item.task.ScheduledAt = time.Now().Add(delay)
	heap.Push(&q.pending, item)
	select {
	case q.notify <- struct{}{}:
	default:
	}
	return nil
}

func (q *DistributedTaskQueue) GetDeadLetterQueue(_ context.Context, limit int) ([]Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if limit <= 0 || limit > len(q.dlq) {
		limit = len(q.dlq)
	}
	out := make([]Task, limit)
	copy(out, q.dlq[:limit])
	return out, nil
}

func (q *DistributedTaskQueue) Stats(_ context.Context) (QueueStats, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	completed, failed := 0, 0
	for _, r := range q.results {
		if r.Success {
			completed++
		} else {
			failed++
		}
	}
	return QueueStats{
		PendingTasks:    q.pending.Len(),
		ProcessingTasks: len(q.processing),
		CompletedTasks:  completed,
		FailedTasks:     failed,
		DLQTasks:        len(q.dlq),
	}, nil
}

// --- TaskWorker ---

type TaskWorker struct {
	queue    TaskQueue
	handlers map[string]TaskHandler
	wg       sync.WaitGroup
	once     sync.Once
	cancel   context.CancelFunc
}

func NewTaskWorker(queue TaskQueue) *TaskWorker {
	return &TaskWorker{queue: queue, handlers: make(map[string]TaskHandler)}
}

func (w *TaskWorker) RegisterHandler(taskType string, h TaskHandler) error {
	w.handlers[taskType] = h
	return nil
}

func (w *TaskWorker) Start(ctx context.Context) error {
	ctx, w.cancel = context.WithCancel(ctx)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			task, err := w.queue.Dequeue(ctx, 500*time.Millisecond)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}
			w.wg.Add(1)
			taskCopy := *task
			go func(t Task) {
				defer w.wg.Done()
				execCtx := ctx
				var cancel context.CancelFunc
				if t.Timeout > 0 {
					execCtx, cancel = context.WithTimeout(ctx, t.Timeout)
					defer cancel()
				}
				// In practice, look up handler by task type embedded in Payload
				w.queue.Complete(execCtx, t.ID, nil)
			}(taskCopy)
		}
	}()
	return nil
}

func (w *TaskWorker) Stop(_ context.Context) error {
	w.once.Do(func() { w.cancel() })
	w.wg.Wait()
	return nil
}

func main() {
	q := NewDistributedTaskQueue()
	ctx := context.Background()

	q.Enqueue(ctx, Task{ID: "t1", Priority: PriorityHigh, MaxRetries: 3})
	q.Enqueue(ctx, Task{ID: "t2", Priority: PriorityLow, MaxRetries: 1})
	q.Enqueue(ctx, Task{ID: "t1", MaxRetries: 3}) // дубль — должен игнорироваться

	task, _ := q.Dequeue(ctx, 1*time.Second)
	fmt.Println("dequeued:", task.ID, "priority:", task.Priority)
	q.Complete(ctx, task.ID, "done")

	task, _ = q.Dequeue(ctx, 1*time.Second)
	q.Fail(ctx, task.ID, fmt.Errorf("something failed"))

	stats, _ := q.Stats(ctx)
	fmt.Printf("stats: %+v\n", stats)
}
