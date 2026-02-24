package main

// Задача: Map-Reduce Framework — master, workers, partitioner, shuffler, fault tolerance.

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type MapFunc func(key, value interface{}) ([]KeyValue, error)
type ReduceFunc func(key interface{}, values []interface{}) (interface{}, error)
type CombineFunc func(key interface{}, values []interface{}) ([]interface{}, error)

type KeyValue struct {
	Key   interface{}
	Value interface{}
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusKilled    TaskStatus = "killed"
)

type TaskType string

const (
	TaskTypeMap    TaskType = "map"
	TaskTypeReduce TaskType = "reduce"
)

type Task struct {
	ID          string
	JobID       string
	Type        TaskType
	SplitID     string
	Partition   int
	Attempt     int
	Status      TaskStatus
	WorkerID    string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       error
}

type InputSplit struct {
	ID       string
	Location string
	Offset   int64
	Length   int64
	Data     []KeyValue
}

type JobConfig struct {
	MaxAttempts          int
	TaskTimeout          time.Duration
	SpeculativeExecution bool
	LocalityPreference   bool
	Compression          bool
}

type Job struct {
	ID          string
	Name        string
	MapFunc     MapFunc
	ReduceFunc  ReduceFunc
	CombineFunc CombineFunc
	Input       []InputSplit
	NumReducers int
	Config      JobConfig
}

type JobStatus struct {
	JobID          string
	Status         TaskStatus
	Progress       float64
	MapProgress    float64
	ReduceProgress float64
	StartedAt      time.Time
	CompletedAt    *time.Time
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	RunningTasks   int
}

type JobResult struct {
	JobID     string
	Success   bool
	Output    []KeyValue
	Duration  time.Duration
	TaskStats TaskStatistics
	Error     error
}

type TaskStatistics struct {
	TotalMapTasks     int
	TotalReduceTasks  int
	FailedMapTasks    int
	FailedReduceTasks int
	AvgMapDuration    time.Duration
	AvgReduceDuration time.Duration
	DataShuffled      int64
	SpeculativeTasks  int
}

type WorkerStatus struct {
	ID             string
	Location       string
	IsAlive        bool
	CurrentTasks   []string
	CPUUsage       float64
	MemoryUsage    int64
	LastHeartbeat  time.Time
	TasksCompleted int64
	TasksFailed    int64
}

// --- HashPartitioner ---

type HashPartitioner struct{}

func (p *HashPartitioner) Partition(key interface{}, numPartitions int) int {
	h := 0
	for _, c := range fmt.Sprintf("%v", key) {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h % numPartitions
}

// --- InMemoryShuffler ---

type InMemoryShuffler struct {
	mu   sync.Mutex
	data map[string]map[string][]KeyValue // jobID -> mapTaskID -> kvs
}

func NewInMemoryShuffler() *InMemoryShuffler {
	return &InMemoryShuffler{data: make(map[string]map[string][]KeyValue)}
}

func (s *InMemoryShuffler) Store(_ context.Context, jobID, mapTaskID string, output []KeyValue) error {
	s.mu.Lock()
	if s.data[jobID] == nil {
		s.data[jobID] = make(map[string][]KeyValue)
	}
	s.data[jobID][mapTaskID] = output
	s.mu.Unlock()
	return nil
}

func (s *InMemoryShuffler) Shuffle(_ context.Context, mapOutput []KeyValue, numReducers int) (map[int][]KeyValue, error) {
	p := &HashPartitioner{}
	partitions := make(map[int][]KeyValue, numReducers)
	for _, kv := range mapOutput {
		partition := p.Partition(kv.Key, numReducers)
		partitions[partition] = append(partitions[partition], kv)
	}
	return partitions, nil
}

func (s *InMemoryShuffler) Fetch(_ context.Context, jobID string, partition int) ([]KeyValue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []KeyValue
	for _, kvs := range s.data[jobID] {
		for _, kv := range kvs {
			result = append(result, kv)
		}
	}
	_ = partition
	return result, nil
}

// --- MapReduceWorker ---

type MapReduceWorker struct {
	id       string
	mu       sync.Mutex
	tasks    []string
	alive    bool
	completed int64
	failed   int64
}

func NewMapReduceWorker(id string) *MapReduceWorker {
	return &MapReduceWorker{id: id, alive: true}
}

func (w *MapReduceWorker) ID() string { return w.id }

func (w *MapReduceWorker) ExecuteMap(_ context.Context, task Task, mapFunc MapFunc) ([]KeyValue, error) {
	w.mu.Lock()
	w.tasks = append(w.tasks, task.ID)
	w.mu.Unlock()
	// task.SplitID would be used to fetch the actual split data
	var result []KeyValue
	// In practice, the worker would read from the InputSplit
	w.mu.Lock()
	w.completed++
	w.mu.Unlock()
	return result, nil
}

func (w *MapReduceWorker) ExecuteReduce(_ context.Context, task Task, reduceFunc ReduceFunc, input []KeyValue) (interface{}, error) {
	// Group by key
	grouped := make(map[interface{}][]interface{})
	for _, kv := range input {
		grouped[kv.Key] = append(grouped[kv.Key], kv.Value)
	}
	var result interface{}
	var err error
	for k, vs := range grouped {
		result, err = reduceFunc(k, vs)
		if err != nil {
			return nil, err
		}
	}
	_ = task
	w.mu.Lock()
	w.completed++
	w.mu.Unlock()
	return result, nil
}

func (w *MapReduceWorker) GetStatus() WorkerStatus {
	w.mu.Lock()
	defer w.mu.Unlock()
	return WorkerStatus{
		ID: w.id, IsAlive: w.alive,
		CurrentTasks: append([]string{}, w.tasks...),
		LastHeartbeat: time.Now(),
		TasksCompleted: w.completed, TasksFailed: w.failed,
	}
}

func (w *MapReduceWorker) Heartbeat(_ context.Context) error { w.alive = true; return nil }

// --- MapReduceMaster (the framework) ---

type jobState struct {
	job         Job
	status      TaskStatus
	startedAt   time.Time
	completedAt *time.Time
	mapTasks    []*Task
	reduceTasks []*Task
	output      []KeyValue
	mu          sync.Mutex
}

type MapReduceMaster struct {
	mu       sync.Mutex
	jobs     map[string]*jobState
	workers  []*MapReduceWorker
	shuffler *InMemoryShuffler
}

func NewMapReduceMaster() *MapReduceMaster {
	return &MapReduceMaster{
		jobs:     make(map[string]*jobState),
		shuffler: NewInMemoryShuffler(),
	}
}

func (m *MapReduceMaster) AddWorker(w *MapReduceWorker) { m.workers = append(m.workers, w) }

func (m *MapReduceMaster) SubmitJob(_ context.Context, job Job) (string, error) {
	js := &jobState{job: job, status: TaskStatusPending, startedAt: time.Now()}
	for i, split := range job.Input {
		js.mapTasks = append(js.mapTasks, &Task{
			ID: fmt.Sprintf("%s-map-%d", job.ID, i), JobID: job.ID,
			Type: TaskTypeMap, SplitID: split.ID, Status: TaskStatusPending,
		})
	}
	for i := 0; i < job.NumReducers; i++ {
		js.reduceTasks = append(js.reduceTasks, &Task{
			ID: fmt.Sprintf("%s-reduce-%d", job.ID, i), JobID: job.ID,
			Type: TaskTypeReduce, Partition: i, Status: TaskStatusPending,
		})
	}
	m.mu.Lock()
	m.jobs[job.ID] = js
	m.mu.Unlock()
	return job.ID, nil
}

func (m *MapReduceMaster) WaitForCompletion(ctx context.Context, jobID string) (*JobResult, error) {
	m.mu.Lock()
	js, ok := m.jobs[jobID]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("job %q not found", jobID)
	}
	start := time.Now()

	// Run map phase
	var allMapOutput []KeyValue
	var mu sync.Mutex
	var wg sync.WaitGroup
	partitioner := &HashPartitioner{}
	for idx, task := range js.mapTasks {
		task := task
		split := js.job.Input[idx]
		wg.Add(1)
		go func() {
			defer wg.Done()
			task.Status = TaskStatusRunning
			var out []KeyValue
			for _, kv := range split.Data {
				kvs, err := js.job.MapFunc(kv.Key, kv.Value)
				if err != nil {
					task.Status = TaskStatusFailed
					return
				}
				// Apply combiner if present
				if js.job.CombineFunc != nil {
					grouped := make(map[interface{}][]interface{})
					for _, mkv := range kvs {
						grouped[mkv.Key] = append(grouped[mkv.Key], mkv.Value)
					}
					kvs = nil
					for k, vs := range grouped {
						combined, _ := js.job.CombineFunc(k, vs)
						for _, v := range combined {
							kvs = append(kvs, KeyValue{Key: k, Value: v})
						}
					}
				}
				out = append(out, kvs...)
			}
			mu.Lock()
			allMapOutput = append(allMapOutput, out...)
			mu.Unlock()
			task.Status = TaskStatusCompleted
		}()
	}
	wg.Wait()

	// Shuffle
	partitions := make(map[int][]KeyValue, js.job.NumReducers)
	for _, kv := range allMapOutput {
		p := partitioner.Partition(kv.Key, js.job.NumReducers)
		partitions[p] = append(partitions[p], kv)
	}

	// Reduce phase
	var output []KeyValue
	for _, task := range js.reduceTasks {
		task.Status = TaskStatusRunning
		kvs := partitions[task.Partition]
		// Sort by key for group-by
		sort.Slice(kvs, func(i, j int) bool {
			return fmt.Sprintf("%v", kvs[i].Key) < fmt.Sprintf("%v", kvs[j].Key)
		})
		grouped := make(map[interface{}][]interface{})
		for _, kv := range kvs {
			grouped[kv.Key] = append(grouped[kv.Key], kv.Value)
		}
		for k, vs := range grouped {
			result, err := js.job.ReduceFunc(k, vs)
			if err != nil {
				task.Status = TaskStatusFailed
				continue
			}
			output = append(output, KeyValue{Key: k, Value: result})
			task.Status = TaskStatusCompleted
		}
	}

	now := time.Now()
	js.status = TaskStatusCompleted
	js.completedAt = &now
	js.output = output

	return &JobResult{
		JobID: jobID, Success: true, Output: output, Duration: time.Since(start),
	}, nil
}

func (m *MapReduceMaster) GetJobStatus(_ context.Context, jobID string) (*JobStatus, error) {
	m.mu.Lock()
	js, ok := m.jobs[jobID]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("job %q not found", jobID)
	}
	js.mu.Lock()
	defer js.mu.Unlock()
	completedMap, totalMap := 0, len(js.mapTasks)
	for _, t := range js.mapTasks {
		if t.Status == TaskStatusCompleted {
			completedMap++
		}
	}
	completedReduce, totalReduce := 0, len(js.reduceTasks)
	for _, t := range js.reduceTasks {
		if t.Status == TaskStatusCompleted {
			completedReduce++
		}
	}
	total := totalMap + totalReduce
	completed := completedMap + completedReduce
	return &JobStatus{
		JobID:          jobID,
		Status:         js.status,
		TotalTasks:     total,
		CompletedTasks: completed,
		StartedAt:      js.startedAt,
		CompletedAt:    js.completedAt,
		Progress:       float64(completed) / float64(total),
	}, nil
}

func (m *MapReduceMaster) CancelJob(_ context.Context, jobID string) error {
	m.mu.Lock()
	js, ok := m.jobs[jobID]
	m.mu.Unlock()
	if ok {
		js.status = TaskStatusKilled
	}
	return nil
}

func main() {
	master := NewMapReduceMaster()
	ctx := context.Background()

	// Word count
	job := Job{
		ID:          "wc-1",
		Name:        "word-count",
		NumReducers: 2,
		Input: []InputSplit{
			{ID: "s1", Data: []KeyValue{
				{Key: "line1", Value: "hello world hello"},
				{Key: "line2", Value: "go go go"},
			}},
		},
		MapFunc: func(key, value interface{}) ([]KeyValue, error) {
			var out []KeyValue
			for _, word := range splitWords(fmt.Sprintf("%v", value)) {
				out = append(out, KeyValue{Key: word, Value: 1})
			}
			return out, nil
		},
		ReduceFunc: func(key interface{}, values []interface{}) (interface{}, error) {
			sum := 0
			for _, v := range values {
				sum += v.(int)
			}
			return sum, nil
		},
	}

	jobID, _ := master.SubmitJob(ctx, job)
	result, _ := master.WaitForCompletion(ctx, jobID)
	fmt.Printf("Word count completed in %v, output:\n", result.Duration)
	sort.Slice(result.Output, func(i, j int) bool {
		return fmt.Sprint(result.Output[i].Key) < fmt.Sprint(result.Output[j].Key)
	})
	for _, kv := range result.Output {
		fmt.Printf("  %v: %v\n", kv.Key, kv.Value)
	}
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\t' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
