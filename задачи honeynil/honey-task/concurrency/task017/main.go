// ЗАДАЧА 17: Group (аналог sync.WaitGroup)
package main

import (
	"reflect"
	"sort"
	"sync"
)

// Group — аналог sync.WaitGroup на каналах.
type Group struct {
	c    chan struct{}
	size int
}

// New создаёт Group для size горутин.
func New(size int) *Group {
	c := make(chan struct{}, size)
	return &Group{c: c, size: size}
}

// Done сигнализирует о завершении одной горутины.
func (s *Group) Done() {
	s.c <- struct{}{}
}

// Wait блокируется до завершения всех size горутин.
func (s *Group) Wait() {
	for i := 0; i < s.size; i++ {
		<-s.c
	}
}

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	n := len(numbers)

	var res []int
	var mu sync.Mutex

	group := New(n)

	for _, num := range numbers {
		go func(num int) {
			defer group.Done()

			mu.Lock()
			res = append(res, num)
			mu.Unlock()
		}(num)
	}

	group.Wait()

	sort.IntSlice(res).Sort()

	if !reflect.DeepEqual(res, numbers) {
		panic("wrong code")
	}
}
