// Package pointers smth task
package pointers

import (
	"fmt"
)

type Person struct {
	name string
	age  int
}

type Counter struct {
	count int
}

type Point struct {
	x, y int
}

type Node struct {
	value int
	next  *Node
}

type Box struct {
	value int
}

type Inner struct {
	value int
}

type Outer struct {
	Inner
	name string
}

type Data struct {
	values []int
}

func (c Counter) increment() {
	c.count++
}

func (c *Counter) incrementPtr() {
	c.count++
}

func (c *Counter) doublePtr() {
	c.count *= 2
}

func (b *Box) setValue(v int) {
	b.value = v
}

const Count = 50

func GetTasks() map[int]func() {
	return map[int]func(){
		1: task1, 2: task2, 3: task3, 4: task4, 5: task5, 6: task6, 7: task7, 8: task8, 9: task9, 10: task10,
		11: task11, 12: task12, 13: task13, 14: task14, 15: task15, 16: task16, 17: task17, 18: task18, 19: task19, 20: task20,
		21: task21, 22: task22, 23: task23, 24: task24, 25: task25, 26: task26, 27: task27, 28: task28, 29: task29, 30: task30,
		31: task31, 32: task32, 33: task33, 34: task34, 35: task35, 36: task36, 37: task37, 38: task38, 39: task39, 40: task40,
		41: task41, 42: task42, 43: task43, 44: task44, 45: task45, 46: task46, 47: task47, 48: task48, 49: task49, 50: task50,
	}
}

// ЗАДАЧА 1: Что выведет?
// OUTPUT:
//
//	10 10   ← x=10, p=&x → *p=10
//	20 20   ← *p=20 меняет x
func task1() {
	x := 10
	p := &x
	fmt.Println(x, *p)
	*p = 20
	fmt.Println(x, *p)
}

// ЗАДАЧА 2: Что выведет?
// OUTPUT:
//
//	<nil>   ← нулевой указатель выводится как <nil>
//	true
func task2() {
	var p *int
	fmt.Println(p)
	fmt.Println(p == nil)
}

// ЗАДАЧА 3: Что выведет?
// OUTPUT:
//
//	0    ← new(int) выделяет память и инициализирует zero value (0 для int)
//	42
func task3() {
	p := new(int)
	fmt.Println(*p)
	*p = 42
	fmt.Println(*p)
}

// ЗАДАЧА 4: Что выведет?
// OUTPUT:
//
//	Alice 30
//	Alice 31   ← p.age = 31 изменяет поле через указатель
func task4() {
	p := &Person{name: "Alice", age: 30}
	fmt.Println(p.name, p.age)
	p.age = 31
	fmt.Println(p.name, p.age)
}

// ЗАДАЧА 5: Что выведет?
// OUTPUT:
//
//	10   ← changeValue(x) — передача по значению, копия; x не меняется
//	20   ← changePointer(&x) — передача указателя; *ptr=20 меняет x
func task5() {
	x := 10
	changeValue(x)
	fmt.Println(x)
	changePointer(&x)
	fmt.Println(x)
}
func changeValue(val int) {
	val = 20
}
func changePointer(ptr *int) {
	*ptr = 20
}

// ЗАДАЧА 6: Что выведет?
// OUTPUT:
//
//	20   ← p1 и p2 указывают на одну переменную x; *p1=20 меняет x; *p2 тоже = 20
func task6() {
	x := 10
	p1 := &x
	p2 := &x
	*p1 = 20
	fmt.Println(*p2)
}

// ЗАДАЧА 7: Что выведет?
// OUTPUT:
//
//	true   ← &x == p: p = &x, поэтому адреса одинаковы
//	true   ← &x == &x: адрес переменной всегда один и тот же
func task7() {
	x := 10
	p := &x
	fmt.Println(&x == p)
	fmt.Println(&x == &x)
}

// ЗАДАЧА 8: Что выведет?
// OUTPUT:
//
//	10 10 10   ← x=10, p=&x → *p=10, pp=&p → **pp=10
//	20         ← **pp=20 меняет x через двойное разыменование
func task8() {
	x := 10
	p := &x
	pp := &p
	fmt.Println(x, *p, **pp)
	**pp = 20
	fmt.Println(x)
}

// ЗАДАЧА 9: Что выведет?
// OUTPUT:
//
//	{0 0}   ← new(Point) возвращает *Point с zero value полями: x=0, y=0
func task9() {
	p := new(Point)
	fmt.Println(*p)
}

// ЗАДАЧА 10: Что выведет?
// OUTPUT:
//
//	false   ← px и py — разные переменные, разные адреса (px != py)
//	true    ← *px = 10 = *py = 10
func task10() {
	x, y := 10, 10
	px := &x
	py := &y
	fmt.Println(px == py)
	fmt.Println(*px == *py)
}

// ЗАДАЧА 11: Что выведет?
// OUTPUT:
//
//	1 2 3
//	← &arr[i] берёт адрес реального элемента массива (не копии переменной цикла)
//	← ptrs[i] указывают на arr[0], arr[1], arr[2]
func task11() {
	arr := [3]int{1, 2, 3}
	var ptrs []*int
	for i := range arr {
		ptrs = append(ptrs, &arr[i])
	}
	for _, p := range ptrs {
		fmt.Print(*p, " ")
	}
	fmt.Println()
}

// ЗАДАЧА 12: Что выведет?
// OUTPUT:
//
//	42   ← Go использует escape analysis: x "убегает" на heap, &x валиден после return
func task12() {
	p := getPointer()
	fmt.Println(*p)
}
func getPointer() *int {
	x := 42
	return &x
}

// ЗАДАЧА 13: Что выведет?
// OUTPUT:
//
//	0   ← value receiver increment: копия, c.count не меняется
//	1   ← pointer receiver incrementPtr: c.count = 0+1 = 1
func task13() {
	c := Counter{count: 0}
	c.increment()
	fmt.Println(c.count)
	c.incrementPtr()
	fmt.Println(c.count)
}

// ЗАДАЧА 14: Что выведет?
// OUTPUT:
//
//	Before panic
//	Recovered from panic: runtime error: invalid memory address or nil pointer dereference
//	← разыменование nil указателя → panic; recover перехватывает
func task14() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	var p *int
	fmt.Println("Before panic")
	fmt.Println(*p)
	fmt.Println("After panic")
}

// ЗАДАЧА 15: Что выведет?
// OUTPUT:
//
//	[100 2 3]   ← p := &arr; p[0] = 100 разыменовывает указатель на массив
func task15() {
	arr := [3]int{1, 2, 3}
	p := &arr
	p[0] = 100
	fmt.Println(arr)
}

// ЗАДАЧА 16: Что выведет?
// OUTPUT:
//
//	20 10   ← swap(&x, &y) через указатели меняет x и y местами
func task16() {
	x, y := 10, 20
	swap(&x, &y)
	fmt.Println(x, y)
}
func swap(a, b *int) {
	*a, *b = *b, *a
}

// ЗАДАЧА 17: Что выведет?
// OUTPUT:
//
//	[1 100 3]   ← p := &s[1]; *p = 100 меняет s[1]
func task17() {
	s := []int{1, 2, 3}
	p := &s[1]
	*p = 100
	fmt.Println(s)
}

// ЗАДАЧА 18: Что выведет?
// OUTPUT:
//
//	1 2   ← n1.next = &n2; n1.next.value = n2.value = 2
func task18() {
	n1 := Node{value: 1}
	n2 := Node{value: 2}
	n1.next = &n2
	fmt.Println(n1.value, n1.next.value)
}

// ЗАДАЧА 19: Что выведет?
// OUTPUT:
//
//	{Alice 30}   ← updatePerson19 изменяет поля через указатель
func task19() {
	p := Person{name: "Bob", age: 25}
	updatePerson19(&p)
	fmt.Println(p)
}
func updatePerson19(p *Person) {
	p.name = "Alice"
	p.age = 30
}

// ЗАДАЧА 20: Что выведет?
// OUTPUT:
//
//	false false   ← b=false, p=&b; *p=false → b=false; оба = false
func task20() {
	b := true
	p := &b
	*p = false
	fmt.Println(b, *p)
}

// ЗАДАЧА 21: Что выведет?
// OUTPUT:
//
//	world world   ← *p = "world" меняет строку s; строка — immutable, но переменная s переназначается
func task21() {
	s := "hello"
	p := &s
	*p = "world"
	fmt.Println(s, *p)
}

// ЗАДАЧА 22: Что выведет?
// OUTPUT:
//
//	2 4 6   ← *p = *p * 2 меняет оригинальные a, b, c через указатели
func task22() {
	a, b, c := 1, 2, 3
	ptrs := []*int{&a, &b, &c}
	for _, p := range ptrs {
		*p = *p * 2
	}
	fmt.Println(a, b, c)
}

// ЗАДАЧА 23: Что выведет?
// OUTPUT:
//
//	hello   ← x — interface{}, p = &x (*interface{}); *p = "hello" меняет x
func task23() {
	var x interface{} = 10
	p := &x
	*p = "hello"
	fmt.Println(x)
}

// ЗАДАЧА 24: Что выведет?
// OUTPUT:
//
//	20   ← d2 := d1 копирует struct, но value — указатель на x; оба d1 и d2 имеют один указатель
func task24() {
	type Data24 struct {
		value *int
	}
	x := 10
	d1 := Data24{value: &x}
	d2 := d1
	*d2.value = 20
	fmt.Println(*d1.value)
}

// ЗАДАЧА 25: Что выведет?
// OUTPUT:
//
//	20   ← m["a"] = &x; *m["a"] = 20 меняет x
func task25() {
	m := make(map[string]*int)
	x := 10
	m["a"] = &x
	*m["a"] = 20
	fmt.Println(x)
}

// ЗАДАЧА 26: Что выведет?
// OUTPUT:
//
//	2   ← (*p)["a"] = 2 разыменовывает указатель на map и меняет элемент
func task26() {
	m := map[string]int{"a": 1}
	p := &m
	(*p)["a"] = 2
	fmt.Println(m["a"])
}

// ЗАДАЧА 27: Что выведет?
// OUTPUT:
//
//	10   ← функциональный литерал; result = f(5) = 5*2 = 10
func task27() {
	f := func(x int) int { return x * 2 }
	result := f(5)
	fmt.Println(result)
}

// ЗАДАЧА 28: Что выведет?
// OUTPUT:
//
//	10   ← c.doublePtr(): c.count=5 * 2 = 10
//	20   ← pc.doublePtr(): pc=&c; c.count=10 * 2 = 20
func task28() {
	c := Counter{count: 5}
	c.doublePtr()
	fmt.Println(c.count)

	pc := &c
	pc.doublePtr()
	fmt.Println(c.count)
}

// ЗАДАЧА 29: Что выведет?
// OUTPUT:
//
//	42   ← p := &ch (*chan int); *p <- 42 отправляет в канал; <-*p читает из канала
func task29() {
	ch := make(chan int, 1)
	p := &ch
	*p <- 42
	fmt.Println(<-*p)
}

// ЗАДАЧА 30: Что выведет?
// OUTPUT:
//
//	true   ← new([]int) создаёт *[]int; *p1 == nil (slice == nil)
//	true   ← new(map[string]int) создаёт *map; *p2 == nil (map == nil)
func task30() {
	p1 := new([]int)
	fmt.Println(*p1 == nil)

	p2 := new(map[string]int)
	fmt.Println(*p2 == nil)
}

// ЗАДАЧА 31: Что выведет?
// OUTPUT:
//
//	3 3 3
//	← Go 1.22+: каждая итерация создаёт новую переменную v → &v уникален для каждой итерации
//	← В Go < 1.22: все &v указывали бы на одну переменную → вывод был бы "3 3 3" (последнее значение)
//	← Начиная с Go 1.22, ответ: 1 2 3 (каждая итерация — своя переменная)
//
// ПРИМЕЧАНИЕ: с Go 1.22 выводит "1 2 3"
func task31() {
	nums := []int{1, 2, 3}
	var ptrs []*int
	for _, v := range nums {
		ptrs = append(ptrs, &v)
	}
	for _, p := range ptrs {
		fmt.Print(*p, " ")
	}
	fmt.Println()
}

// ЗАДАЧА 32: Что выведет?
// OUTPUT:
//
//	1 2 3
//	← &nums[i] берёт адрес элемента slice, независимо от версии Go
func task32() {
	nums := []int{1, 2, 3}
	var ptrs []*int
	for i := range nums {
		ptrs = append(ptrs, &nums[i])
	}
	for _, p := range ptrs {
		fmt.Print(*p, " ")
	}
	fmt.Println()
}

// ЗАДАЧА 33: Что выведет?
// OUTPUT:
//
//	[1 2 3 4]   ← appendToSlicePtr принимает *[]int; *s = append(*s, val) меняет оригинальный slice
func task33() {
	s := []int{1, 2, 3}
	appendToSlicePtr(&s, 4)
	fmt.Println(s)
}
func appendToSlicePtr(s *[]int, val int) {
	*s = append(*s, val)
}

// ЗАДАЧА 34: Что выведет?
// OUTPUT:
//
//	42   ← горутина меняет x через указатель; done <- true синхронизирует
func task34() {
	x := 0
	done := make(chan bool)
	go func(p *int) {
		*p = 42
		done <- true
	}(&x)
	<-done
	fmt.Println(x)
}

// ЗАДАЧА 35: Что выведет?
// OUTPUT:
//
//	20   ← p1 и p2 указывают на одну struct b; p1.value=20 меняет b → p2.value=20
func task35() {
	type Box struct{ value int }
	b := Box{value: 10}
	p1, p2 := &b, &b
	p1.value = 20
	fmt.Println(p2.value)
}

// ЗАДАЧА 36: Что выведет?
// OUTPUT:
//
//	10 20   ← автоматическое разыменование указателя на struct: p.x, p.y
func task36() {
	p := &Point{x: 10, y: 20}
	fmt.Println(p.x, p.y)
}

// ЗАДАЧА 37: Что выведет?
// OUTPUT:
//
//	100   ← escape analysis: x escape на heap; &x валиден после return
func task37() {
	p := getIntPtr(100)
	fmt.Println(*p)
}
func getIntPtr(x int) *int {
	return &x
}

// ЗАДАЧА 38: Что выведет?
// OUTPUT:
//
//	{1 2}   ← *p разыменовывает указатель на struct; %v выводит поля
func task38() {
	p := &Point{x: 1, y: 2}
	fmt.Println(*p)
}

// ЗАДАЧА 39: Что выведет?
// OUTPUT:
//
//	10   ← b.setValue(10): Go автоматически берёт адрес для pointer receiver
//	15   ← (&b).setValue(15): явно
func task39() {
	b := Box{value: 5}
	b.setValue(10)
	fmt.Println(b.value)
	(&b).setValue(15)
	fmt.Println(b.value)
}

// ЗАДАЧА 40: Что выведет?
// OUTPUT:
//
//	[1 100 3]   ← p := &arr[1]; *p = 100 меняет arr[1]
func task40() {
	arr := [3]int{1, 2, 3}
	p := &arr[1]
	*p = 100
	fmt.Println(arr)
}

// ЗАДАЧА 41: Что выведет?
// OUTPUT:
//
//	true   ← n.next == nil (не инициализирован)
func task41() {
	n := Node{value: 1, next: nil}
	fmt.Println(n.next == nil)
}

// ЗАДАЧА 42: Что выведет?
// OUTPUT:
//
//	1   ← defer вычисляет аргументы СРАЗУ при регистрации (не при выполнении)
//	    ← *getPtr(&x) вычисляется при x=1 → результат 1
//	    ← x=2 происходит ПОСЛЕ регистрации defer, поэтому не влияет на аргумент
//	    ← КЛЮЧЕВОЕ: defer f(args...) — args вычисляются немедленно
func task42() {
	x := 1
	defer fmt.Println(*getPtr(&x))
	x = 2
}
func getPtr(p *int) *int {
	return p
}

// ЗАДАЧА 43: Что выведет?
// OUTPUT:
//
//	42   ← getValue() возвращает 42; p := &x; *p = x = 42
func task43() {

	x := getValue()
	p := &x
	fmt.Println(*p)
}
func getValue() int {
	return 42
}

// ЗАДАЧА 44: Что выведет?
// OUTPUT:
//
//	1 2 3   ← связный список n1→n2→n3
func task44() {
	n3 := &Node{value: 3, next: nil}
	n2 := &Node{value: 2, next: n3}
	n1 := &Node{value: 1, next: n2}
	fmt.Println(n1.value, n1.next.value, n1.next.next.value)
}

// ЗАДАЧА 45: Что выведет?
// OUTPUT:
//
//	10   ← p = &x; *p = 10
//	20   ← p = &y (переназначение p); *p = 20
func task45() {
	x, y := 10, 20
	p := &x
	fmt.Println(*p)
	p = &y
	fmt.Println(*p)
}

// ЗАДАЧА 46: Что выведет?
// OUTPUT:
//
//	100   ← p := &o.Inner берёт адрес embedded Inner; p.value=100 меняет o.Inner.value
func task46() {
	o := Outer{Inner: Inner{value: 42}, name: "test"}
	p := &o.Inner
	p.value = 100
	fmt.Println(o.value)
}

// ЗАДАЧА 47: Что выведет?
// OUTPUT:
//
//	Charlie   ← map хранит *Person47; m[1].name = "Charlie" меняет через указатель
func task47() {
	type Person47 struct{ name string }
	m := map[int]*Person47{
		1: {name: "Alice"},
		2: {name: "Bob"},
	}
	m[1].name = "Charlie"
	fmt.Println(m[1].name)
}

// ЗАДАЧА 48: Что выведет?
// OUTPUT:
//
//	[1 100 3 4 5 6 7 8 9 10]
//	← p := &s[1] берёт адрес s[1] в исходном backing array
//	← append(...) реаллоцирует s (новый backing array)
//	← *p = 200 пишет в СТАРЫЙ backing array, который больше не используется s
//	← s не изменился (остался с исходными данными после реаллокации)
//	← КЛЮЧЕВОЕ: после реаллокации p — "висячий" указатель
func task48() {
	s := []int{1, 2, 3}
	p := &s[1]
	*p = 100
	s = append(s, 4, 5, 6, 7, 8, 9, 10)
	*p = 200
	fmt.Println(s)
}

// ЗАДАЧА 49: Что выведет?
// OUTPUT:
//
//	[100 2 3]    ← modifyDataValue49(d) — передача по значению; d.values — slice (общий backing array)
//	             ← d.values[0]=100 меняет оригинальный slice!
//	[100 200 3]  ← modifyDataPointer49(&d) — указатель; d.values[1]=200
func task49() {
	d := Data{values: []int{1, 2, 3}}
	modifyDataValue49(d)
	fmt.Println(d.values)
	modifyDataPointer49(&d)
	fmt.Println(d.values)
}
func modifyDataValue49(d Data) {
	d.values[0] = 100
}
func modifyDataPointer49(d *Data) {
	d.values[1] = 200
}

// ЗАДАЧА 50: Что выведет?
// OUTPUT:
//
//	10 0
//	← createOnStack() возвращает указатель на локальную переменную (escape на heap)
//	← createOnHeap() использует new(int) → zero value = 0
//	← оба варианта корректны: Go сам решает stack/heap через escape analysis
func task50() {
	p1 := createOnStack()
	p2 := createOnHeap()
	fmt.Println(*p1, *p2)
}
func createOnStack() *int {
	x := 10
	return &x
}
func createOnHeap() *int {
	return new(int)
}
