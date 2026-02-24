// ЗАДАЧА 13: ConcurrentSortHead
package main

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func main() {
	f1 := "aaa\nddd\n"
	f2 := "bbb\neee\n"
	f3 := "ccc\nfff\n"

	files := []io.Reader{
		strings.NewReader(f1),
		strings.NewReader(f2),
		strings.NewReader(f3),
	}

	rows, err := ConcurrentSortHead(4, files...)
	fmt.Println(rows)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(rows, []string{"aaa", "bbb", "ccc", "ddd"}) {
		panic("wrong code")
	}
}

// lineItem — элемент для heap: строка и индекс источника.
type lineItem struct {
	line   string
	source int
}

// minHeap — мин-куча строк для k-way merge.
type minHeap []lineItem

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].line < h[j].line }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(lineItem)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// ConcurrentSortHead конкурентно читает строки из ридеров и возвращает m первых.
func ConcurrentSortHead(m int, files ...io.Reader) ([]string, error) {
	type chanResult struct {
		ch  <-chan string
		idx int
	}

	// Конкурентно читаем строки из каждого файла в отдельный канал
	chans := make([]<-chan string, len(files))
	for i, f := range files {
		ch := make(chan string)
		chans[i] = ch
		go func(r io.Reader, out chan<- string) {
			defer close(out)
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				out <- scanner.Text()
			}
		}(f, ch)
	}

	// K-way merge с мин-кучей
	h := &minHeap{}
	heap.Init(h)

	// Читаем первую строку из каждого файла
	type chState struct {
		ch   <-chan string
		done bool
	}
	states := make([]chState, len(chans))
	for i, ch := range chans {
		states[i] = chState{ch: ch}
		if line, ok := <-ch; ok {
			heap.Push(h, lineItem{line, i})
		} else {
			states[i].done = true
		}
	}

	var result []string
	for h.Len() > 0 && len(result) < m {
		item := heap.Pop(h).(lineItem)
		result = append(result, item.line)
		// Читаем следующую строку из того же источника
		if line, ok := <-states[item.source].ch; ok {
			heap.Push(h, lineItem{line, item.source})
		}
	}
	return result, nil
}
