package main

// Задача: SAGA Orchestrator — транзакции с компенсациями.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TransactionStatus string

const (
	StatusPending      TransactionStatus = "pending"
	StatusInProgress   TransactionStatus = "in_progress"
	StatusCompleted    TransactionStatus = "completed"
	StatusFailed       TransactionStatus = "failed"
	StatusCompensating TransactionStatus = "compensating"
	StatusCompensated  TransactionStatus = "compensated"
)

type Step struct {
	Name       string
	Action     func(ctx context.Context, data interface{}) (interface{}, error)
	Compensate func(ctx context.Context, data interface{}) error
	Timeout    time.Duration
}

type Transaction struct {
	ID          string
	Steps       []Step
	Status      TransactionStatus
	Data        map[string]interface{}
	StartedAt   time.Time
	CompletedAt *time.Time
	Error       error
}

type TransactionEvent struct {
	TxID      string
	StepName  string
	EventType string
	Data      interface{}
	Timestamp time.Time
}

type SagaOrchestrator interface {
	Execute(ctx context.Context, tx *Transaction) error
	GetStatus(ctx context.Context, txID string) (TransactionStatus, error)
	Compensate(ctx context.Context, txID string) error
	Resume(ctx context.Context, txID string) error
}

type TransactionLog interface {
	LogStepStarted(ctx context.Context, txID, stepName string, data interface{}) error
	LogStepCompleted(ctx context.Context, txID, stepName string, result interface{}) error
	LogStepFailed(ctx context.Context, txID, stepName string, err error) error
	LogCompensationStarted(ctx context.Context, txID, stepName string) error
	GetTransactionState(ctx context.Context, txID string) (*Transaction, error)
}

// --- InMemoryTransactionLog ---

type InMemoryTransactionLog struct {
	mu     sync.RWMutex
	states map[string]*Transaction
	events map[string][]TransactionEvent
}

func NewInMemoryTransactionLog() *InMemoryTransactionLog {
	return &InMemoryTransactionLog{
		states: make(map[string]*Transaction),
		events: make(map[string][]TransactionEvent),
	}
}

func (l *InMemoryTransactionLog) store(txID, stepName, evType string, data interface{}) {
	l.events[txID] = append(l.events[txID], TransactionEvent{
		TxID: txID, StepName: stepName, EventType: evType, Data: data, Timestamp: time.Now(),
	})
}

func (l *InMemoryTransactionLog) LogStepStarted(_ context.Context, txID, stepName string, data interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store(txID, stepName, "started", data)
	return nil
}
func (l *InMemoryTransactionLog) LogStepCompleted(_ context.Context, txID, stepName string, result interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store(txID, stepName, "completed", result)
	return nil
}
func (l *InMemoryTransactionLog) LogStepFailed(_ context.Context, txID, stepName string, err error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store(txID, stepName, "failed", err)
	return nil
}
func (l *InMemoryTransactionLog) LogCompensationStarted(_ context.Context, txID, stepName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store(txID, stepName, "compensating", nil)
	return nil
}
func (l *InMemoryTransactionLog) GetTransactionState(_ context.Context, txID string) (*Transaction, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	tx, ok := l.states[txID]
	if !ok {
		return nil, fmt.Errorf("transaction %q not found", txID)
	}
	return tx, nil
}

func (l *InMemoryTransactionLog) SaveTx(tx *Transaction) {
	l.mu.Lock()
	l.states[tx.ID] = tx
	l.mu.Unlock()
}

// --- SagaOrchestratorImpl ---

type SagaOrchestratorImpl struct {
	log *InMemoryTransactionLog
}

func NewSagaOrchestrator(log *InMemoryTransactionLog) *SagaOrchestratorImpl {
	return &SagaOrchestratorImpl{log: log}
}

func (o *SagaOrchestratorImpl) Execute(ctx context.Context, tx *Transaction) error {
	tx.Status = StatusInProgress
	tx.StartedAt = time.Now()
	o.log.SaveTx(tx)

	completed := -1
	for i, step := range tx.Steps {
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}
		o.log.LogStepStarted(stepCtx, tx.ID, step.Name, tx.Data)

		result, err := step.Action(stepCtx, tx.Data)
		if err != nil {
			o.log.LogStepFailed(stepCtx, tx.ID, step.Name, err)
			tx.Status = StatusFailed
			tx.Error = err
			o.log.SaveTx(tx)

			// compensate completed steps in reverse
			o.compensateUpTo(ctx, tx, completed)
			return err
		}
		tx.Data[step.Name+"_result"] = result
		o.log.LogStepCompleted(stepCtx, tx.ID, step.Name, result)
		completed = i
	}

	now := time.Now()
	tx.Status = StatusCompleted
	tx.CompletedAt = &now
	o.log.SaveTx(tx)
	return nil
}

func (o *SagaOrchestratorImpl) compensateUpTo(ctx context.Context, tx *Transaction, upTo int) {
	tx.Status = StatusCompensating
	o.log.SaveTx(tx)
	for i := upTo; i >= 0; i-- {
		step := tx.Steps[i]
		if step.Compensate == nil {
			continue
		}
		o.log.LogCompensationStarted(ctx, tx.ID, step.Name)
		step.Compensate(ctx, tx.Data)
	}
	tx.Status = StatusCompensated
	o.log.SaveTx(tx)
}

func (o *SagaOrchestratorImpl) GetStatus(_ context.Context, txID string) (TransactionStatus, error) {
	tx, err := o.log.GetTransactionState(context.Background(), txID)
	if err != nil {
		return "", err
	}
	return tx.Status, nil
}

func (o *SagaOrchestratorImpl) Compensate(ctx context.Context, txID string) error {
	tx, err := o.log.GetTransactionState(ctx, txID)
	if err != nil {
		return err
	}
	o.compensateUpTo(ctx, tx, len(tx.Steps)-1)
	return nil
}

func (o *SagaOrchestratorImpl) Resume(ctx context.Context, txID string) error {
	tx, err := o.log.GetTransactionState(ctx, txID)
	if err != nil {
		return err
	}
	return o.Execute(ctx, tx)
}

func main() {
	log := NewInMemoryTransactionLog()
	orch := NewSagaOrchestrator(log)

	tx := &Transaction{
		ID:   "order-123",
		Data: make(map[string]interface{}),
		Steps: []Step{
			{
				Name: "reserve-inventory",
				Action: func(ctx context.Context, data interface{}) (interface{}, error) {
					fmt.Println("✓ reserving inventory")
					return "inventory-reserved", nil
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					fmt.Println("↩ releasing inventory")
					return nil
				},
			},
			{
				Name: "charge-payment",
				Action: func(ctx context.Context, data interface{}) (interface{}, error) {
					fmt.Println("✗ payment failed")
					return nil, fmt.Errorf("payment declined")
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					fmt.Println("↩ refunding payment")
					return nil
				},
			},
		},
	}

	err := orch.Execute(context.Background(), tx)
	fmt.Println("execute err:", err)
	status, _ := orch.GetStatus(context.Background(), "order-123")
	fmt.Println("final status:", status)
}
