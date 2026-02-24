// Package maps some task
package maps

import (
	"fmt"
	"sort"
)

type Person struct {
	name string
	age  int
}

type Point struct {
	x, y int
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
//	true     ← var m map[string]int объявлена, но не инициализирована → nil
//	false    ← m2 := map[string]int{} инициализирована, не nil
//	0 0      ← len(nil map) == 0, len(пустой map) == 0
func task1() {
	var m1 map[string]int
	m2 := map[string]int{}
	fmt.Println(m1 == nil)
	fmt.Println(m2 == nil)
	fmt.Println(len(m1), len(m2))
}

// ЗАДАЧА 2: Что выведет?
// OUTPUT:
//
//	0        ← чтение из nil map не паникует, возвращает zero value типа (int → 0)
func task2() {
	var m map[string]int
	value := m["key"]
	fmt.Println(value)
}

// ЗАДАЧА 3: Что произойдет?
// OUTPUT:
//
//	Recovered: assignment to entry in nil map
//	← запись в nil map вызывает панику; recover её перехватывает
//
//lint:ignore SA5000
func task3() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()
	var m map[string]int
	m["key"] = 42
	fmt.Println("This won't print")
}

// ЗАДАЧА 4: Что выведет?
// OUTPUT:
//
//	0 true   ← m["a"]=0, ключ существует (ok=true)
//	0 false  ← m["c"] отсутствует: zero value=0, ok=false
func task4() {
	m := map[string]int{"a": 0, "b": 2}
	v1, ok1 := m["a"]
	v2, ok2 := m["c"]
	fmt.Println(v1, ok1)
	fmt.Println(v2, ok2)
}

// ЗАДАЧА 5: Что выведет?
// OUTPUT:
//
//	No panic!  ← delete на nil map НЕ паникует
//	0
func task5() {
	var m map[string]int
	delete(m, "key")
	fmt.Println("No panic!")
	fmt.Println(len(m))
}

// ЗАДАЧА 6: Что выведет?
// OUTPUT:
//
//	map[a:1 b:2]  ← delete несуществующего ключа "c" — тихо игнорируется
//	2
func task6() {
	m := map[string]int{"a": 1, "b": 2}
	delete(m, "c")
	fmt.Println(m)
	fmt.Println(len(m))
}

// ЗАДАЧА 7: Что выведет?
// OUTPUT:
//
//	map[a:2 b:3]  ← map — ссылочный тип! m2 := m1 копирует указатель, не данные
//	map[a:2 b:3]  ← оба указывают на одну хеш-таблицу
func task7() {
	m1 := map[string]int{"a": 1}
	m2 := m1
	m2["a"] = 2
	m2["b"] = 3
	fmt.Println(m1)
	fmt.Println(m2)
}

// ЗАДАЧА 8: Что выведет?
// OUTPUT:
//
//	map[a:100 b:200]  ← map передаётся по ссылке, изменения в функции видны снаружи
func task8() {
	m := map[string]int{"a": 1}
	modifyMap(m)
	fmt.Println(m)
}
func modifyMap(m map[string]int) {
	m["a"] = 100
	m["b"] = 200
}

// ЗАДАЧА 9: Что выведет?
// OUTPUT:
//
//	map[a:1]  ← переприсвоение m внутри функции не меняет оригинал (копируется дескриптор)
func task9() {
	m := map[string]int{"a": 1}
	reassignMap(m)
	fmt.Println(m)
}
func reassignMap(m map[string]int) {
	m = map[string]int{"x": 999}
}

// ЗАДАЧА 10: Что выведет?
// OUTPUT (порядок НЕДЕТЕРМИНИРОВАН, один из вариантов):
//
//	c:3 a:1 b:2
//	← итерация по map в Go происходит в случайном порядке
func task10() {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	for k, v := range m {
		fmt.Print(k, ":", v, " ")
	}
	fmt.Println()
}

// ЗАДАЧА 11: Что выведет?
// OUTPUT (порядок недетерминирован, "d" может попасть или нет):
//
//	a b c d  ← добавление ключа во время итерации допустимо; новый ключ может быть посещён или нет
//	map[a:1 b:2 c:3 d:4]
func task11() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	for k := range m {
		if k == "b" {
			m["d"] = 4
		}
		fmt.Print(k, " ")
	}
	fmt.Println("\n", m)
}

// ЗАДАЧА 12: Что выведет?
// OUTPUT:
//
//	map[]  ← удаление элементов во время итерации безопасно в Go
//	0
func task12() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	for k := range m {
		delete(m, k)
	}
	fmt.Println(m)
	fmt.Println(len(m))
}

// ЗАДАЧА 13: Что выведет?
// OUTPUT:
//
//	map[1:true 2:false 3:true 4:true]
//	false true false  ← set[2]=false (ключ есть), set[4]=true, set[5] — zero value bool = false
func task13() {
	set := map[int]bool{1: true, 2: true, 3: true}
	set[2] = false
	set[4] = true
	fmt.Println(set)
	fmt.Println(set[2], set[4], set[5])
}

// ЗАДАЧА 14: Что выведет?
// OUTPUT:
//
//	A         ← m[{1,2}] = "A"
//	          ← m[{5,6}] отсутствует, zero value string = ""
func task14() {
	m := map[Point]string{
		{1, 2}: "A",
		{3, 4}: "B",
	}
	p := Point{1, 2}
	fmt.Println(m[p])
	fmt.Println(m[Point{5, 6}])
}

// ЗАДАЧА 15: Что произойдет?
// OUTPUT:
//
//	Slices нельзя использовать как ключи map!
//	Ключи должны быть comparable
//	← []int не является comparable (содержит slices), компилятор выдаёт ошибку
func task15() {
	// Раскомментируй чтобы увидеть ошибку компиляции:
	//m := map[[]int]string{}
	fmt.Println("Slices нельзя использовать как ключи map!")
	fmt.Println("Ключи должны быть comparable")
}

// ЗАДАЧА 16: Что выведет?
// OUTPUT:
//
//	map[a:[1 2 3 4] b:[4 5] c:[1]]
//	← append(m["a"], 4) добавляет в slice по ключу "a"
//	← append(m["c"], 1): m["c"]==nil slice, append создаёт новый slice [1]
func task16() {
	m := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5},
	}
	m["a"] = append(m["a"], 4)
	m["c"] = append(m["c"], 1)
	fmt.Println(m)
}

// ЗАДАЧА 17: Что выведет?
// OUTPUT:
//
//	map[user1:map[score:150] user2:map[score:200]]
func task17() {
	m := map[string]map[string]int{
		"user1": {"score": 100},
		"user2": {"score": 200},
	}
	m["user1"]["score"] = 150
	fmt.Println(m)
}

// ЗАДАЧА 18: Что произойдет?
// OUTPUT:
//
//	Panic: assignment to entry in nil map
//	← m["user1"] не инициализирован (nil), попытка записи в него паникует
func task18() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic:", r)
		}
	}()
	m := map[string]map[string]int{}
	m["user1"]["score"] = 100
	fmt.Println("Won't print")
}

// ЗАДАЧА 19: Что выведет?
// OUTPUT:
//
//	map[user1:map[score:100]]
//	← вложенный map инициализирован явно перед записью
func task19() {
	m := map[string]map[string]int{}
	m["user1"] = map[string]int{}
	m["user1"]["score"] = 100
	fmt.Println(m)
}

// ЗАДАЧА 20: Что выведет?
// OUTPUT:
//
//	0   ← make с capacity не влияет на len; len всегда = количество элементов
//	2
func task20() {
	m := make(map[string]int, 10)
	fmt.Println(len(m))
	m["a"] = 1
	m["b"] = 2
	fmt.Println(len(m))
}

// ЗАДАЧА 21: Что выведет?
// OUTPUT:
//
//	(ничего) ← закомментированный код не компилируется: map нельзя сравнивать с ==
//	← Maps не являются comparable, поэтому m1 == m2 вызвало бы ошибку компиляции
func task21() {
	//m1 := map[string]int{"a": 1}
	//m2 := map[string]int{"a": 1}
	//fmt.Println(m1 == m2)
}

// ЗАДАЧА 22: Что выведет?
// OUTPUT:
//
//	map[a:2 b:5]
//	← m["a"]++ работает на zero value (0++→1, 1++→2); m["b"]+=5 → 5
func task22() {
	m := make(map[string]int)
	m["a"]++
	m["a"]++
	m["b"] += 5
	fmt.Println(m)
}

// ЗАДАЧА 23: Что выведет?
// OUTPUT:
//
//	{Alice 31}
//	← нельзя напрямую менять поле в map-value struct; нужно извлечь, изменить, положить обратно
func task23() {
	m := map[string]Person{
		"alice": {"Alice", 30},
		"bob":   {"Bob", 25},
	}

	p := m["alice"]
	p.age = 31
	m["alice"] = p
	fmt.Println(m["alice"])
}

// ЗАДАЧА 24: Что выведет?
// OUTPUT:
//
//	&{Alice 31}
//	← при хранении *Person в map можно напрямую менять поля через указатель
func task24() {
	m := map[string]*Person{
		"alice": {"Alice", 30},
		"bob":   {"Bob", 25},
	}
	m["alice"].age = 31
	fmt.Println(m["alice"])
}

// ЗАДАЧА 25: Что выведет?
// OUTPUT:
//
//	Нельзя, потому что элементы могут перемещаться при росте map
//	1
//	← &m["a"] — ошибка компиляции: нельзя брать адрес map-элемента (рантайм может переместить bucket)
func task25() {
	m := map[string]int{"a": 1}
	//p := &m["a"]
	//fmt.Println(p)
	fmt.Println("Нельзя, потому что элементы могут перемещаться при росте map")
	fmt.Println(m["a"])
}

// ЗАДАЧА 26: Что выведет?
// OUTPUT:
//
//	m1: map[a:1 b:2]   ← m1 не изменился (ручное копирование)
//	m2: map[a:100 b:2]
func task26() {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := make(map[string]int)
	for k, v := range m1 {
		m2[k] = v
	}
	m2["a"] = 100
	fmt.Println("m1:", m1)
	fmt.Println("m2:", m2)
}

// ЗАДАЧА 27: Что выведет?
// OUTPUT (детерминировано, т.к. count ограничен):
//
//	6
//	map[1:1 2:2 3:3 11:1 12:2 13:3]
//	← во время итерации добавляются ключи k+10; добавленные ключи не будут посещены повторно
func task27() {
	m := map[int]int{1: 1, 2: 2, 3: 3}
	count := 0
	for k := range m {
		if count < 5 {
			m[k+10] = k
			count++
		}
	}
	fmt.Println(len(m))
	fmt.Println(m)
}

// ЗАДАЧА 28: Что выведет?
// OUTPUT:
//
//	[a b c]  ← ключи отсортированы через sort.Strings
func task28() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println(keys)
}

// ЗАДАЧА 29: Что выведет?
// OUTPUT:
//
//	Done - no panic on nil map range
//	← итерация по nil map — безопасна, тело цикла просто не выполняется
func task29() {
	var m map[string]int
	for k, v := range m {
		fmt.Println(k, v)
	}
	fmt.Println("Done - no panic on nil map range")
}

// ЗАДАЧА 30: Что выведет?
// OUTPUT:
//
//	42
//	hello
//	52   ← type assertion m["int"].(int) извлекает конкретный тип; 42+10=52
func task30() {
	m := map[string]interface{}{
		"int":    42,
		"string": "hello",
		"bool":   true,
	}
	fmt.Println(m["int"])
	fmt.Println(m["string"])
	val := m["int"].(int)
	fmt.Println(val + 10)
}

// ЗАДАЧА 31: Что выведет?
// OUTPUT:
//
//	Found: 2
//	map[a:1 c:3]
func task31() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	if v, ok := m["b"]; ok {
		fmt.Println("Found:", v)
		delete(m, "b")
	}
	fmt.Println(m)
}

// ЗАДАЧА 32: Что выведет?
// OUTPUT:
//
//	[map[a:100] map[b:2 x:999] map[c:3]]
func task32() {
	slice := []map[string]int{
		{"a": 1},
		{"b": 2},
		{"c": 3},
	}
	slice[0]["a"] = 100
	slice[1]["x"] = 999
	fmt.Println(slice)
}

// ЗАДАЧА 33: Что выведет?
// OUTPUT:
//
//	Added 10000 items
//	Len after delete: 0
//	← delete не освобождает память сразу; len возвращает количество элементов = 0
func task33() {
	m := make(map[int]int)
	for i := 0; i < 10000; i++ {
		m[i] = i
	}
	fmt.Println("Added 10000 items")
	for i := 0; i < 10000; i++ {
		delete(m, i)
	}
	fmt.Println("Len after delete:", len(m))
}

// ЗАДАЧА 34: Что выведет?
// OUTPUT:
//
//	map[a:3 b:2]  ← перезапись существующего ключа
func task34() {
	m := map[string]int{"a": 1, "b": 2}
	m["a"] = 3
	fmt.Println(m)
}

// ЗАДАЧА 35: Что выведет?
// OUTPUT:
//
//	true   ← set["exists"] = true
//	false  ← set["notreally"] = false (ключ есть, но значение false)
//	false  ← set["missing"] — zero value bool = false
//	← ВАЖНО: нельзя отличить "ключ отсутствует" от "ключ=false" без двух-значного получения
func task35() {
	m := map[string]bool{
		"exists":    true,
		"also":      true,
		"notreally": false,
	}
	fmt.Println(m["exists"])
	fmt.Println(m["notreally"])
	fmt.Println(m["missing"])
}

// ЗАДАЧА 36: Что выведет?
// OUTPUT:
//
//	Alice   ← m[Key{"user", 1}] = "Alice"
//	        ← m[Key{"user", 3}] отсутствует, zero value string = ""
func task36() {
	type Key struct {
		name string
		id   int
	}
	m := map[Key]string{
		{"user", 1}: "Alice",
		{"user", 2}: "Bob",
	}
	fmt.Println(m[Key{"user", 1}])
	fmt.Println(m[Key{"user", 3}])
}

// ЗАДАЧА 37: Что выведет?
// OUTPUT (недетерминировано):
//
//	Last key: <один из ключей: "a", "b" или "c">
//	← последний ключ при итерации — случайный
func task37() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	var lastKey string
	for k := range m {
		lastKey = k
	}
	fmt.Println("Last key:", lastKey)
}

// ЗАДАЧА 38: Что выведет?
// OUTPUT:
//
//	0
//	map[x:100]
//	← после очистки map продолжает быть usable
func task38() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	for k := range m {
		delete(m, k)
	}
	fmt.Println(len(m))
	m["x"] = 100
	fmt.Println(m)
}

// ЗАДАЧА 39: Что выведет?
// OUTPUT:
//
//	map[a:1 b:2]
//	map[a:1 b:2]
//	← p := &m — указатель на map; (*p)["b"] = 2 меняет оригинальный map
func task39() {
	m := map[string]int{"a": 1}
	p := &m
	(*p)["b"] = 2
	fmt.Println(m)
	fmt.Println(*p)
}

// ЗАДАЧА 40: Что выведет?
// OUTPUT:
//
//	map[x:999]
//	← changeMapPtr принимает *map, разыменовывает и заменяет весь map новым
func task40() {
	m := map[string]int{"a": 1}
	changeMapPtr(&m)
	fmt.Println(m)
}
func changeMapPtr(m *map[string]int) {
	*m = map[string]int{"x": 999}
}

// ЗАДАЧА 41: Что выведет?
// OUTPUT:
//
//	A
//	A
//	← массив [2]int является comparable и может быть ключом map
func task41() {
	m := map[[2]int]string{
		{1, 2}: "A",
		{3, 4}: "B",
	}
	key := [2]int{1, 2}
	fmt.Println(m[key])
	fmt.Println(m[[2]int{1, 2}])
}

// ЗАДАЧА 42: Что выведет?
// OUTPUT:
//
//	100
//	3
//	← User{1, "Alice"} и User{1, "Different"} — разные ключи (разное поле Name)
func task42() {
	type User struct {
		ID   int
		Name string
	}
	m := map[User]int{
		{1, "Alice"}: 100,
		{2, "Bob"}:   200,
	}
	u := User{1, "Alice"}
	fmt.Println(m[u])
	m[User{1, "Different"}] = 300
	fmt.Println(len(m))
}

// ЗАДАЧА 43: Что выведет?
// OUTPUT:
//
//	map[a:2 b:2 c:100]
func task43() {
	m := map[string]int{"a": 1, "b": 2}
	if _, ok := m["a"]; ok {
		m["a"] = m["a"] * 2
	}
	if _, ok := m["c"]; !ok {
		m["c"] = 100
	}
	fmt.Println(m)
}

// ЗАДАЧА 44: Что выведет?
// OUTPUT:
//
//	[a b c]    ← ключи отсортированы
//	[1 2 3]    ← значения отсортированы независимо от ключей
//	← ВАЖНО: порядок ключей и значений независим после sort; не отражает пары k→v
func task44() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	var keys []string
	var values []int
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	sort.Strings(keys)
	sort.Ints(values)
	fmt.Println(keys)
	fmt.Println(values)
}

// ЗАДАЧА 45: Что выведет?
// OUTPUT:
//
//	map[apple:3 banana:2 cherry:1]
func task45() {
	words := []string{"apple", "banana", "apple", "cherry", "banana", "apple"}
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}
	fmt.Println(counts)
}

// ЗАДАЧА 46: Что выведет?
// OUTPUT:
//
//	map[fruit:[1 2 4] veg:[3]]
func task46() {
	type Item struct {
		category string
		value    int
	}
	items := []Item{
		{"fruit", 1},
		{"fruit", 2},
		{"veg", 3},
		{"fruit", 4},
	}
	groups := make(map[string][]int)
	for _, item := range items {
		groups[item.category] = append(groups[item.category], item.value)
	}
	fmt.Println(groups)
}

// ЗАДАЧА 47: Что выведет?
// OUTPUT:
//
//	42
//	99
//	← map может хранить каналы; отправка/получение работают через обычный синтаксис
func task47() {
	m := map[string]chan int{
		"ch1": make(chan int, 1),
		"ch2": make(chan int, 1),
	}
	m["ch1"] <- 42
	m["ch2"] <- 99
	fmt.Println(<-m["ch1"])
	fmt.Println(<-m["ch2"])
}

// ЗАДАЧА 48: Что выведет?
// OUTPUT:
//
//	v1: 0 v2: 0    ← m["a"]=0 (zero value int), m["c"] не существует → zero value = 0
//	ok1: true ok2: false
func task48() {
	m := map[string]int{"a": 0, "b": 5}
	v1 := m["a"]
	v2 := m["c"]
	fmt.Println("v1:", v1, "v2:", v2)
	_, ok1 := m["a"]
	_, ok2 := m["c"]
	fmt.Println("ok1:", ok1, "ok2:", ok2)
}

// ЗАДАЧА 49: Что выведет?
// OUTPUT:
//
//	Before: 3
//	After: 2
//	map[1:one 3:three]
//	← delete(m, 5) — несуществующий ключ, ошибки нет
func task49() {
	m := map[int]string{1: "one", 2: "two", 3: "three"}
	fmt.Println("Before:", len(m))
	delete(m, 2)
	delete(m, 5)
	fmt.Println("After:", len(m))
	fmt.Println(m)
}

// ЗАДАЧА 50: Что выведет?
// OUTPUT:
//
//	map[a:1 b:3 c:4]
//	← слияние двух map: m2["b"]=3 перезаписывает m1["b"]=2
func task50() {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	result := make(map[string]int)
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	fmt.Println(result)
}
