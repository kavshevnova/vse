// ЗАДАЧА 15: RingBuffer
package main

import (
	"fmt"
	"reflect"
)

// ringBuffer — кольцевой буфер фиксированного размера.
// При записи в полный буфер перезаписывает самые старые элементы.
type ringBuffer struct {
	data  []int
	size  int
	head  int // индекс для чтения
	tail  int // индекс для записи
	count int
	done  bool
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		data: make([]int, size),
		size: size,
	}
}

func (b *ringBuffer) write(v int) {
	b.data[b.tail] = v
	b.tail = (b.tail + 1) % b.size
	if b.count < b.size {
		b.count++
	} else {
		// Перезаписываем: сдвигаем head вперёд
		b.head = (b.head + 1) % b.size
	}
}

func (b *ringBuffer) close() {
	b.done = true
}

func (b *ringBuffer) read() (v int, ok bool) {
	if b.count == 0 {
		return 0, false
	}
	v = b.data[b.head]
	b.head = (b.head + 1) % b.size
	b.count--
	return v, true
}

func main() {
	buff := newRingBuffer(3)

	for i := 1; i <= 6; i++ {
		buff.write(i)
	}

	buff.close()

	res := make([]int, 0)
	for {
		if v, ok := buff.read(); ok {
			res = append(res, v)
		} else {
			break
		}
	}

	if !reflect.DeepEqual(res, []int{4, 5, 6}) {
		panic(fmt.Sprintf("wrong code, res is %v", res))
	}
	fmt.Println(res) // [4 5 6]
}
