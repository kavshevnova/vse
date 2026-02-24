package main

// Задача: реализовать функцию Pipe — читает из Producer батчами, буферизует до MaxItems, пишет в Consumer.

import "fmt"

const MaxItems = 9999

type Producer interface {
	Next() (items []any, cookie int, err error)
	Commit(cookie int) error
}

type Consumer interface {
	Process(items []any) error
}

// Pipe читает данные из Producer, буферизует до MaxItems и передаёт в Consumer.
// После успешной обработки батча фиксирует cookie в Producer.
func Pipe(p Producer, c Consumer) error {
	var (
		buffer  []any
		cookies []int
	)

	flush := func() error {
		if len(buffer) == 0 {
			return nil
		}
		if err := c.Process(buffer); err != nil {
			return err
		}
		// Коммитим все cookie строго в порядке получения
		for _, cookie := range cookies {
			if err := p.Commit(cookie); err != nil {
				return err
			}
		}
		buffer = buffer[:0]
		cookies = cookies[:0]
		return nil
	}

	for {
		items, cookie, err := p.Next()
		if err != nil {
			// Флашим оставшееся перед выходом
			if ferr := flush(); ferr != nil {
				return ferr
			}
			return err
		}

		buffer = append(buffer, items...)
		cookies = append(cookies, cookie)

		if len(buffer) >= MaxItems {
			if err := flush(); err != nil {
				return err
			}
		}
	}
}

// --- Демо ---

type mockProducer struct {
	batches [][]any
	pos     int
}

func (p *mockProducer) Next() ([]any, int, error) {
	if p.pos >= len(p.batches) {
		return nil, 0, fmt.Errorf("EOF")
	}
	batch := p.batches[p.pos]
	cookie := p.pos
	p.pos++
	return batch, cookie, nil
}

func (p *mockProducer) Commit(cookie int) error {
	fmt.Printf("committed cookie %d\n", cookie)
	return nil
}

type mockConsumer struct{}

func (c *mockConsumer) Process(items []any) error {
	fmt.Printf("processed %d items\n", len(items))
	return nil
}

func main() {
	p := &mockProducer{batches: [][]any{
		{1, 2, 3},
		{4, 5},
		{6},
	}}
	c := &mockConsumer{}
	err := Pipe(p, c)
	fmt.Println("done:", err)
}
