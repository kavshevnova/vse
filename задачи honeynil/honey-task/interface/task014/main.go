package main

// Задача: Scheduler — одноразовые и повторяющиеся задачи, cron-like расписание.

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ScheduledTask interface {
	ID() string
	Run() error
	Cancel()
}

type Schedule interface {
	Next(after time.Time) time.Time
}

type Scheduler interface {
	ScheduleOnce(delay time.Duration, task func() error) (ScheduledTask, error)
	ScheduleRepeat(interval time.Duration, task func() error) (ScheduledTask, error)
	ScheduleWithSchedule(schedule Schedule, task func() error) (ScheduledTask, error)
	Cancel(taskID string) error
	Start() error
	Stop() error
}

// --- scheduledTask ---

type scheduledTask struct {
	id     string
	fn     func() error
	cancel context.CancelFunc
}

func (t *scheduledTask) ID() string   { return t.id }
func (t *scheduledTask) Run() error   { return t.fn() }
func (t *scheduledTask) Cancel()      { t.cancel() }

// --- PeriodicSchedule ---

type PeriodicSchedule struct{ Interval time.Duration }

func (s *PeriodicSchedule) Next(after time.Time) time.Time { return after.Add(s.Interval) }

// --- CronSchedule (упрощённый: просто обёртка над периодом) ---

type CronSchedule struct{ interval time.Duration }

func NewCronSchedule(interval time.Duration) *CronSchedule { return &CronSchedule{interval: interval} }
func (c *CronSchedule) Next(after time.Time) time.Time     { return after.Add(c.interval) }

// --- SimpleScheduler ---

type SimpleScheduler struct {
	mu    sync.Mutex
	tasks map[string]*scheduledTask
	ctx   context.Context
	cancel context.CancelFunc
}

func newID() string { return fmt.Sprintf("%016x", rand.Int63()) }

func NewSimpleScheduler() *SimpleScheduler { return &SimpleScheduler{tasks: make(map[string]*scheduledTask)} }

func (s *SimpleScheduler) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return nil
}

func (s *SimpleScheduler) Stop() error {
	s.cancel()
	return nil
}

func (s *SimpleScheduler) register(t *scheduledTask) {
	s.mu.Lock()
	s.tasks[t.id] = t
	s.mu.Unlock()
}

func (s *SimpleScheduler) ScheduleOnce(delay time.Duration, fn func() error) (ScheduledTask, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	t := &scheduledTask{id: newID(), fn: fn, cancel: cancel}
	s.register(t)
	go func() {
		select {
		case <-time.After(delay):
			t.fn()
		case <-ctx.Done():
		}
	}()
	return t, nil
}

func (s *SimpleScheduler) ScheduleRepeat(interval time.Duration, fn func() error) (ScheduledTask, error) {
	return s.ScheduleWithSchedule(&PeriodicSchedule{Interval: interval}, fn)
}

func (s *SimpleScheduler) ScheduleWithSchedule(sched Schedule, fn func() error) (ScheduledTask, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	t := &scheduledTask{id: newID(), fn: fn, cancel: cancel}
	s.register(t)
	go func() {
		next := sched.Next(time.Now())
		for {
			select {
			case <-time.After(time.Until(next)):
				fn()
				next = sched.Next(time.Now())
			case <-ctx.Done():
				return
			}
		}
	}()
	return t, nil
}

func (s *SimpleScheduler) Cancel(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tasks[taskID]; ok {
		t.cancel()
		delete(s.tasks, taskID)
	}
	return nil
}

func main() {
	scheduler := NewSimpleScheduler()
	scheduler.Start()

	count := 0
	t, _ := scheduler.ScheduleRepeat(100*time.Millisecond, func() error {
		count++
		fmt.Println("tick", count)
		return nil
	})

	time.Sleep(350 * time.Millisecond)
	t.Cancel()
	fmt.Println("total ticks:", count)
	scheduler.Stop()
}
