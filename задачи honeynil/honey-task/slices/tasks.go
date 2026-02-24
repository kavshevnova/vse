// Package slices some tasks
package slices

import (
	"fmt"
	"reflect"
)

const Count = 50

func GetTasks() map[int]func() {
	return map[int]func(){
		1: a, 2: b, 3: c, 4: d, 5: e, 6: f, 7: g, 8: h, 9: i, 10: j,
		11: k, 12: l, 13: m, 14: n, 15: o, 16: p, 17: q, 18: r, 19: s, 20: t,
		21: u, 22: v, 23: w, 24: x, 25: y, 26: z, 27: aa, 28: ab, 29: ac, 30: ad,
		31: ae, 32: af, 33: ag, 34: ah, 35: ai, 36: aj, 37: ak, 38: al, 39: am, 40: an,
		41: ao, 42: ap, 43: aq, 44: ar, 45: as, 46: at, 47: au, 48: av, 49: aw, 50: ax,
	}
}

// ЗАДАЧА 1 Что выведет?
// OUTPUT:
//
//	x =  [] len:  0 cap:  0
//	x =  [0] len:  1 cap:  1
//	x =  [0 1] len:  2 cap:  2
//	x =  [0 1 2] len:  3 cap:  4       ← cap удвоился: 2→4
//	y =  [0 1 2 3] len:  4 cap:  4
//	z =  [0 1 2 4] len:  4 cap:  4     ← y и z оба append к x (len=3, cap=4)
//	[0 1 2 4] [0 1 2 4]                ← y[3] перезаписан z[3]=4, т.к. разделяют backing array!
//
// КЛЮЧЕВОЕ: y и z добавляются в одно место backing array → z перезаписывает y[3]
func a() {
	x := []int{}
	fmt.Println("x = ", x, "len: ", len(x), "cap: ", cap(x))
	x = append(x, 0)
	fmt.Println("x = ", x, "len: ", len(x), "cap: ", cap(x))
	x = append(x, 1)
	fmt.Println("x = ", x, "len: ", len(x), "cap: ", cap(x))
	x = append(x, 2)
	fmt.Println("x = ", x, "len: ", len(x), "cap: ", cap(x))
	y := append(x, 3)
	fmt.Println("y = ", y, "len: ", len(y), "cap: ", cap(y))
	z := append(x, 4)
	fmt.Println("z = ", z, "len: ", len(z), "cap: ", cap(z))
	fmt.Println(y, z)
}

// ЗАДАЧА 2: Что выведет?
// OUTPUT:
//
//	true false   ← var x []int → nil; y := []int{} → не nil (инициализирован)
//	0 0          ← len обоих = 0
//	0 0          ← cap обоих = 0
func b() {
	var x []int
	y := []int{}
	fmt.Println(x == nil, y == nil)
	fmt.Println(len(x), len(y))
	fmt.Println(cap(x), cap(y))
}

// ЗАДАЧА 3: Что выведет?
// OUTPUT:
//
//	0 5    ← make([]int, 0, 5): len=0, cap=5
//	3 5    ← append 1,2,3: len=3, cap=5 (не превысили cap)
//	6 10   ← append 4,5,6: len=6 > 5, перераспределение: cap=10
func c() {
	x := make([]int, 0, 5)
	fmt.Println(len(x), cap(x))
	x = append(x, 1, 2, 3)
	fmt.Println(len(x), cap(x))
	x = append(x, 4, 5, 6)
	fmt.Println(len(x), cap(x))
}

// ЗАДАЧА 4: Что выведет?
// OUTPUT:
//
//	[1 100 3 4 5]   ← y := x[1:3] разделяет backing array с x; y[0]=x[1]
//	[100 3]
func d() {
	x := []int{1, 2, 3, 4, 5}
	y := x[1:3]
	y[0] = 100
	fmt.Println(x)
	fmt.Println(y)
}

// ЗАДАЧА 5: Что выведет?
// OUTPUT:
//
//	[1 2 3 100 5]   ← y := x[1:3], cap(y)=4; append(y,100) записывает в x[3]
//	[2 3 100]
func e() {
	x := []int{1, 2, 3, 4, 5}
	y := x[1:3]
	y = append(y, 100)
	fmt.Println(x)
	fmt.Println(y)
}

// ЗАДАЧА 6: Что выведет?
// OUTPUT:
//
//	[100 2 3]   ← y := x — копирует дескриптор (ptr,len,cap), но backing array общий
func f() {
	x := []int{1, 2, 3}
	y := x
	y[0] = 100
	fmt.Println(x)
}

// ЗАДАЧА 7: Что выведет?
// OUTPUT:
//
//	[1 2]   ← copy(dst, src) копирует min(len(dst),len(src)) элементов = 2
func g() {
	x := []int{1, 2, 3, 4, 5}
	y := make([]int, 2)
	copy(y, x)
	fmt.Println(y)
}

// ЗАДАЧА 8: Что выведет?
// OUTPUT:
//
//	[2 3 3]   ← copy(x, x[1:]) → копирует x[1],x[2] → x[0],x[1]: [2,3,3]
func h() {
	x := []int{1, 2, 3}
	copy(x, x[1:])
	fmt.Println(x)
}

// ЗАДАЧА 9: Что выведет?
// OUTPUT:
//
//	[2 4 6]   ← range возвращает копию значения v, но x[i] изменяется по индексу
func i() {
	x := []int{1, 2, 3}
	for i, v := range x {
		x[i] = v * 2
	}
	fmt.Println(x)
}

// ЗАДАЧА 10: Что выведет?
// OUTPUT:
//
//	[{Alice} {Bob}]   ← range копирует v; изменение копии v.name не влияет на slice
func j() {
	type Person struct{ name string }
	x := []Person{{"Alice"}, {"Bob"}}
	for _, v := range x {
		v.name = "Changed"
	}
	fmt.Println(x)
}

// ЗАДАЧА 11: Что выведет?
// OUTPUT:
//
//	[1 2 3 100 5]   ← y := x[:3] (cap=5); append(y,100) пишет в x[3]
//	[1 2 3 100]
func k() {
	x := []int{1, 2, 3, 4, 5}
	y := x[:3]
	z := append(y, 100)
	fmt.Println(x)
	fmt.Println(z)
}

// ЗАДАЧА 12: Что выведет?
// OUTPUT:
//
//	[1 0 0 2]    ← make([]int,3,5): [0,0,0] + append(2) = [0,0,0,2]; x[0]=1 → [1,0,0,2]
//	4 5
func l() {
	x := make([]int, 3, 5)
	x[0] = 1
	x = append(x, 2)
	fmt.Println(x)
	fmt.Println(len(x), cap(x))
}

// ЗАДАЧА 13: Что выведет?
// OUTPUT:
//
//	[1 100 3 4 5]   ← slice от массива arr разделяет память; x[0]=arr[1]
//	[100 3]
func m() {
	arr := [5]int{1, 2, 3, 4, 5}
	x := arr[1:3]
	x[0] = 100
	fmt.Println(arr)
	fmt.Println(x)
}

// ЗАДАЧА 14: Что выведет?
// OUTPUT:
//
//	[1]   ← append к nil slice работает нормально
func n() {
	var x []int
	x = append(x, 1)
	fmt.Println(x)
}

// ЗАДАЧА 15: Что выведет?
// OUTPUT:
//
//	[1 2] [1 2 3] [1 2 4] [1 2 5]
//	← x имеет cap=2=len; каждый append создаёт новый backing array → независимы
func o() {
	x := []int{1, 2}
	y := append(x, 3)
	z := append(x, 4)
	w := append(x, 5)
	fmt.Println(x, y, z, w)
}

// ЗАДАЧА 16: Что выведет?
// OUTPUT:
//
//	Нужно написать функцию для сравнения 3 3
//	true   ← reflect.DeepEqual сравнивает slice по значению
func p() {
	x := []int{1, 2, 3}
	y := []int{1, 2, 3}
	fmt.Println("Нужно написать функцию для сравнения", len(x), len(y))
	fmt.Println(reflect.DeepEqual(x, y))
}

// ЗАДАЧА 17: Что выведет?
// OUTPUT:
//
//	2 2   ← x[:2:2] — трёхиндексный срез: len=2, cap=2 (cap ограничен до 2)
func q() {
	x := make([]int, 3, 10)
	y := x[:2:2]
	fmt.Println(len(y), cap(y))
}

// ЗАДАЧА 18: Что выведет?
// OUTPUT:
//
//	[1 2 3 4 5] | 5 6
//	← append(x, y...) добавляет все элементы y; cap удваивается: 2→4, потом 4→6? нет: 2+3=5>4 → новый cap=6
func r() {
	x := []int{1, 2}
	y := []int{3, 4, 5}
	x = append(x, y...)
	fmt.Println(x, "|", len(x), cap(x))
}

// ЗАДАЧА 19: Что выведет?
// OUTPUT:
//
//	[1 2 4 5]   ← удаление элемента по индексу i=2: append(x[:2], x[3:]...)
func s() {
	x := []int{1, 2, 3, 4, 5}
	i := 2
	x = append(x[:i], x[i+1:]...)
	fmt.Println(x)
}

// ЗАДАЧА 20: Что выведет?
// OUTPUT:
//
//	[5 4 3 2 1]   ← разворот slice на месте
func t() {
	x := []int{1, 2, 3, 4, 5}
	for i := 0; i < len(x)/2; i++ {
		x[i], x[len(x)-1-i] = x[len(x)-1-i], x[i]
	}
	fmt.Println(x)
}

// ЗАДАЧА 21: Что выведет?
// OUTPUT:
//
//	4    ← make([]int, 0, 4): cap=4
//	4    ← append 1,2,3,4: ровно вмещается, cap=4
//	8    ← append 5: переполнение, cap удваивается: 4→8
func u() {
	x := make([]int, 0, 4)
	fmt.Println(cap(x))
	x = append(x, 1, 2, 3, 4)
	fmt.Println(cap(x))
	x = append(x, 5)
	fmt.Println(cap(x))
}

// ЗАДАЧА 22: Что выведет?
// OUTPUT:
//
//	[4 5]   ← x[2:] = [3,4,5]; y[1:] = [4,5]
func v() {
	x := []int{1, 2, 3, 4, 5}
	y := x[2:]
	z := y[1:]
	fmt.Println(z)
}

// ЗАДАЧА 23: Что выведет?
// OUTPUT:
//
//	[2 100]    ← y[1]=x[2], z[0]=x[2]; изменение y[1]=100 меняет и z[0]
//	[100 4]
func w() {
	x := []int{1, 2, 3, 4, 5}
	y := x[1:3]
	z := x[2:4]
	y[1] = 100
	fmt.Println(y)
	fmt.Println(z)
}

// ЗАДАЧА 24: Что выведет?
// OUTPUT:
//
//	[0 2 4 6 8]
//	5 8   ← cap при росте nil slice: 1→2→4→8
func x() {
	var result []int
	for i := 0; i < 5; i++ {
		result = append(result, i*2)
	}
	fmt.Println(result)
	fmt.Println(len(result), cap(result))
}

// ЗАДАЧА 25: Что выведет?
// OUTPUT:
//
//	[1 2 3 4 5]   ← вставка 3 на позицию i=2: append(x[:2], append([]int{3}, x[2:]...)...)
func y() {
	x := []int{1, 2, 4, 5}
	i := 2
	x = append(x[:i], append([]int{3}, x[i:]...)...)
	fmt.Println(x)
}

// ЗАДАЧА 26: Что выведет?
// OUTPUT:
//
//	0 8           ← result := x[:0]: len=0, cap=8 (разделяет backing array с x)
//	[2 4 6 8]     ← фильтр чётных чисел
//	[2 4 6 8 5 6 7 8]  ← x перезаписан! result и x делят backing array
func z() {
	x := []int{1, 2, 3, 4, 5, 6, 7, 8}
	result := x[:0]
	fmt.Println(len(result), cap(result))
	for _, v := range x {
		if v%2 == 0 {
			result = append(result, v)
		}
	}
	fmt.Println(result)
	fmt.Println(x)
}

// ЗАДАЧА 27: Что выведет?
// OUTPUT:
//
//	[[100 2] [3 4]]   ← вложенные slice разделяют backing array; x[0][0] изменяется
func aa() {
	x := [][]int{{1, 2}, {3, 4}}
	x[0][0] = 100
	fmt.Println(x)
}

// ЗАДАЧА 28: Что выведет?
// OUTPUT:
//
//	100 2    ← *x[0] = 100 меняет переменную a
//	100 2
func ab() {
	a, b := 1, 2
	x := []*int{&a, &b}
	*x[0] = 100
	fmt.Println(a, b)
	fmt.Println(*x[0], *x[1])
}

// ЗАДАЧА 29: Что выведет?
// OUTPUT:
//
//	[0 1 2]   ← append к nil slice
func ac() {
	var x []int
	for i := 0; i < 3; i++ {
		x = append(x, i)
	}
	fmt.Println(x)
}

// ЗАДАЧА 30: Что выведет?
// OUTPUT:
//
//	[]    ← x[1:1] — пустой срез (len=0)
//	0
func ad() {
	x := []int{1, 2, 3, 4, 5}
	y := x[1:1]
	fmt.Println(y)
	fmt.Println(len(y))
}

// ЗАДАЧА 31: Что выведет?
// OUTPUT:
//
//	[100 2 3]   ← slice передаётся по значению (дескриптор), но backing array общий
func ae() {
	x := []int{1, 2, 3}
	modifySlice(x)
	fmt.Println(x)
}
func modifySlice(s []int) {
	s[0] = 100
}

// ЗАДАЧА 32: Что выведет?
// OUTPUT:
//
//	[1 2 3] | 3 3   ← appendToSlice создаёт новый backing array (cap=3 → 6), не влияет на x
func af() {
	x := []int{1, 2, 3}
	appendToSlice(x)
	fmt.Println(x, "|", len(x), cap(x))
}
func appendToSlice(s []int) {
	s = append(s, 4)
}

// ЗАДАЧА 33: Что выведет?
// OUTPUT:
//
//	hello   ← строки неизменяемы; []byte(s) создаёт копию
//	Hello
func ag() {
	s := "hello"
	x := []byte(s)
	x[0] = 'H'
	fmt.Println(s)
	fmt.Println(string(x))
}

// ЗАДАЧА 34: Что выведет?
// OUTPUT:
//
//	12   ← len("привет") в байтах: кириллица — 2 байта на символ, 6 символов × 2 = 12
//	6    ← len([]rune) = 6 символов
func ah() {
	s := "привет"
	x := []rune(s)
	fmt.Println(len(s))
	fmt.Println(len(x))
}

// ЗАДАЧА 35: Что выведет?
// OUTPUT:
//
//	len:  0 cap:  0
//	len: 1 cap: 4    ← Go 1.18+ новая стратегия роста cap: 0→1→2→4 (не обязательно x2)
//	len: 2 cap: 4
//	len: 3 cap: 4
//	len: 4 cap: 4
//	len: 5 cap: 8
func ai() {
	x := []int{}

	fmt.Println("len: ", len(x), "cap: ", cap(x))
	for i := 0; i < 5; i++ {
		x = append(x, i)
		fmt.Println("len:", len(x), "cap:", cap(x))
	}
}

// ЗАДАЧА 36: Что выведет?
// OUTPUT:
//
//	[1 2 3 4]
//	4 4   ← append([]int{1}, x...) создаёт новый slice с len=4, cap=4
func aj() {
	x := []int{2, 3, 4}
	x = append([]int{1}, x...)
	fmt.Println(x)
	fmt.Println(len(x), cap(x))
}

// ЗАДАЧА 37: Что выведет?
// OUTPUT:
//
//	[]       ← x = x[:0] уменьшает len до 0, данные не теряются, но невидимы
//	0 5      ← cap остаётся 5
func ak() {
	x := []int{1, 2, 3, 4, 5}
	x = x[:0]
	fmt.Println(x)
	fmt.Println(len(x), cap(x))
}

// ЗАДАЧА 38: Что выведет?
// OUTPUT:
//
//	[1 2 1 2 3]   ← copy(x[2:], x[:3]): копирует x[0..2] в x[2..4]; copy обрабатывает перекрытие
func al() {
	x := []int{1, 2, 3, 4, 5}
	copy(x[2:], x[:3])
	fmt.Println(x)
}

// ЗАДАЧА 39: Что выведет?
// OUTPUT:
//
//	1 9   ← y := x[1:2] из slice с cap=10: len(y)=1, cap(y)=10-1=9
func am() {
	x := make([]int, 3, 10)
	x[0], x[1], x[2] = 1, 2, 3
	y := x[1:2]
	fmt.Println(len(y), cap(y))
}

// ЗАДАЧА 40: Что выведет?
// OUTPUT:
//
//	x:  [1 2 4] | len:  3 | cap:  4
//	y:  [1 2 3] | len:  3 | cap:  4
//	← y := x[0:2:2] ограничивает cap=2; append(y,3) создаёт НОВЫЙ backing array (cap=4)
//	← x = append(x, 4) пишет в исходный backing array (x[2]=4)
//	← y и x теперь независимы (разные backing arrays)
func an() {
	x := make([]int, 2, 4)
	x[0], x[1] = 1, 2
	y := x[0:2:2]
	y = append(y, 3)
	x = append(x, 4)
	fmt.Println("x: ", x, "|", "len: ", len(x), "|", "cap: ", cap(x))
	fmt.Println("y: ", y, "|", "len: ", len(y), "|", "cap: ", cap(y))
}

// ЗАДАЧА 41: Что выведет?
// OUTPUT:
//
//	[1 2 3]     ← x не изменён (len=3=cap=3, нет места)
//	[1 2 3 4]   ← y и z: оба append создают новые backing arrays
//	[1 2 3 5]
func ao() {
	x := []int{1, 2, 3}
	y := x
	z := x
	y = append(y, 4)
	z = append(z, 5)
	fmt.Println(x)
	fmt.Println(y)
	fmt.Println(z)
}

// ЗАДАЧА 42: Что выведет?
// OUTPUT:
//
//	Done   ← range по nil slice — тело цикла не выполняется, паники нет
func ap() {
	var x []int
	for i, v := range x {
		fmt.Println(i, v)
	}
	fmt.Println("Done")
}

// ЗАДАЧА 43: Что выведет?
// OUTPUT:
//
//	true   ← var x []int == nil
//	false  ← y := []int{} != nil
//	true   ← len(x) == 0
//	true   ← len(y) == 0
func aq() {
	var x []int
	y := []int{}
	fmt.Println(x == nil)
	fmt.Println(y == nil)
	fmt.Println(len(x) == 0)
	fmt.Println(len(y) == 0)
}

// ЗАДАЧА 44: Что выведет?
// OUTPUT:
//
//	[1 2 3 4]   ← append([]int{3}, 4) = [3,4]; append(x, [3,4]...) = [1,2,3,4]
func ar() {
	x := []int{1, 2}
	x = append(x, append([]int{3}, 4)...)
	fmt.Println(x)
}

// ЗАДАЧА 45: Что выведет?
// OUTPUT:
//
//	len:  3 cap:  3
//	[1 2 3 0 1 2] 6 6
//	← range захватывает len=3 в начале; цикл выполняется 3 раза (i=0,1,2)
//	← каждый append добавляет i: 0,1,2 → [1,2,3,0,1,2]
func as() {
	x := []int{1, 2, 3}
	fmt.Println("len: ", len(x), "cap: ", cap(x))
	for i := range x {
		x = append(x, i)
		if i > 5 {
			break
		}
	}
	fmt.Println(x, len(x), cap(x))
}

// ЗАДАЧА 46: Что выведет?
// OUTPUT:
//
//	[1 2 3 0 0 0]   ← y := x[:6]: расширяем len до 6 (cap=10 позволяет); элементы 3-5 = zero value = 0
func at() {
	x := make([]int, 3, 10)
	x[0], x[1], x[2] = 1, 2, 3
	y := x[:6]
	fmt.Println(y)
}

// ЗАДАЧА 47: Что выведет?
// OUTPUT:
//
//	[1 2 3 4 5 6 7 8]   ← p := &x[1] берёт адрес ОРИГИНАЛЬНОГО элемента x[1]
//	0x...               ← append реаллоцирует x (новый backing array); p теперь указывает на СТАРЫЙ backing array
//	← *p = 100 меняет старый backing array, НЕ новый x; x не изменился
//
// КЛЮЧЕВОЕ: после realloc указатель p становится "висячим" (dangling) указателем на старый массив
func au() {
	x := []int{1, 2, 3}
	p := &x[1]
	x = append(x, 4, 5, 6, 7, 8)
	*p = 100
	fmt.Println(x)
	fmt.Println(p)
}

// ЗАДАЧА 48: Что выведет?
// OUTPUT:
//
//	[1 2 100 4 5]   ← z := x[2:3]; z[0]=x[2]; изменение z[0]=100 меняет x[2]
//	[100 4]         ← y := x[2:4] разделяет backing array, y[0]=x[2]=100
//	[100]
func av() {
	x := []int{1, 2, 3, 4, 5}
	y := x[2:4]
	z := x[2:3]
	z[0] = 100
	fmt.Println(x)
	fmt.Println(y)
	fmt.Println(z)
}

// ЗАДАЧА 49: Что выведет?
// OUTPUT:
//
//	[1 2 3]   ← удаление дубликатов из отсортированного slice (алгоритм двух указателей)
func aw() {
	x := []int{1, 1, 2, 2, 3, 3}
	j := 0
	for i := 1; i < len(x); i++ {
		if x[i] != x[j] {
			j++
			x[j] = x[i]
		}
	}
	result := x[:j+1]
	fmt.Println(result)
}

// ЗАДАЧА 50: Что выведет?
// OUTPUT:
//
//	[1 2 3 4 5] 5
//	[2 3]
//	2 3             ← y := x[1:3:4]: len=2, cap=3 (ограничен до 4-1=3)
//	[1 2 3 100 5]   ← append(y,100) пишет в x[3] (не реаллоцирует т.к. cap=3>len=2)
//	[2 3 100]
func ax() {
	x := make([]int, 5, 10)
	for i := range x {
		x[i] = i + 1
	}
	fmt.Println(x, len(x))
	y := x[1:3:4]
	fmt.Println(y)
	fmt.Println(len(y), cap(y))
	y = append(y, 100)
	fmt.Println(x)
	fmt.Println(y)
}
