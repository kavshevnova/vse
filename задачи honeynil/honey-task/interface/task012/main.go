package main

// Задача: Stream — Filter, Map, Reduce, Collect, ForEach, Count.

import "fmt"

type Stream interface {
	Filter(predicate func(interface{}) bool) Stream
	Map(transform func(interface{}) interface{}) Stream
	Reduce(initial interface{}, accumulator func(interface{}, interface{}) interface{}) interface{}
	Collect() []interface{}
	ForEach(consumer func(interface{}))
	Count() int
}

// --- SliceStream ---

type SliceStream struct {
	items []interface{}
}

func NewSliceStream(items []interface{}) *SliceStream {
	return &SliceStream{items: items}
}

func (s *SliceStream) Filter(pred func(interface{}) bool) Stream {
	result := make([]interface{}, 0)
	for _, item := range s.items {
		if pred(item) {
			result = append(result, item)
		}
	}
	return &SliceStream{items: result}
}

func (s *SliceStream) Map(transform func(interface{}) interface{}) Stream {
	result := make([]interface{}, len(s.items))
	for i, item := range s.items {
		result[i] = transform(item)
	}
	return &SliceStream{items: result}
}

func (s *SliceStream) Reduce(initial interface{}, acc func(interface{}, interface{}) interface{}) interface{} {
	result := initial
	for _, item := range s.items {
		result = acc(result, item)
	}
	return result
}

func (s *SliceStream) Collect() []interface{} {
	out := make([]interface{}, len(s.items))
	copy(out, s.items)
	return out
}

func (s *SliceStream) ForEach(consumer func(interface{})) {
	for _, item := range s.items {
		consumer(item)
	}
}

func (s *SliceStream) Count() int { return len(s.items) }

// --- ChannelStream ---

type ChannelStream struct {
	items []interface{}
}

func NewChannelStream(ch <-chan interface{}) *ChannelStream {
	var items []interface{}
	for v := range ch {
		items = append(items, v)
	}
	return &ChannelStream{items: items}
}

func (c *ChannelStream) Filter(pred func(interface{}) bool) Stream  { return NewSliceStream(c.items).Filter(pred) }
func (c *ChannelStream) Map(transform func(interface{}) interface{}) Stream { return NewSliceStream(c.items).Map(transform) }
func (c *ChannelStream) Reduce(init interface{}, acc func(interface{}, interface{}) interface{}) interface{} {
	return NewSliceStream(c.items).Reduce(init, acc)
}
func (c *ChannelStream) Collect() []interface{} { return c.items }
func (c *ChannelStream) ForEach(f func(interface{})) { NewSliceStream(c.items).ForEach(f) }
func (c *ChannelStream) Count() int { return len(c.items) }

func main() {
	items := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	stream := NewSliceStream(items)

	// Чётные числа, умноженные на 2
	result := stream.
		Filter(func(v interface{}) bool { return v.(int)%2 == 0 }).
		Map(func(v interface{}) interface{} { return v.(int) * 2 }).
		Collect()
	fmt.Println(result) // [4 8 12 16 20]

	// Сумма
	sum := NewSliceStream(items).Reduce(0, func(acc, v interface{}) interface{} {
		return acc.(int) + v.(int)
	})
	fmt.Println(sum) // 55
}
