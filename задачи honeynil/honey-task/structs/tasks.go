// Package structs some tasks
package structs

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type Person struct {
	Name string
	age  int
}

type Employee struct {
	name   string
	salary int
}

type Counter struct {
	value int
}

type Product struct {
	ID    int
	name  string
	Price float64
}

type Data struct {
	Value int    `json: "value"`
	Name  string `json:"name"`
}

type Config struct {
	Host string `json:"host"`
	port int    `json:"port"`
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	data    string `json:"data"`
}

type Address struct {
	City    string
	Country string
}

type Contact struct {
	Email string
	Phone string
}

type User struct {
	Address
	Contact
	Name string
}

type Base struct {
	ID   int
	Name string
}

type Extended struct {
	Base
	Name  string
	Extra string
}

type Inner struct {
	Value int
}

type Outer struct {
	*Inner
	Label string
}

type Node struct {
	Value int
	Next  *Node
}

type Point struct {
	X, Y int
}

type Metrics struct {
	Count int
}

type Stats struct {
	Metrics
	Total int
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
//	Bob   ← p.Name до вызова
//	Bob   ← changePersonPtr создаёт новый *Person, но не меняет оригинальный указатель p
//	← КЛЮЧЕВОЕ: p передаётся по значению (копия указателя); p = &Person{...} меняет только локальную копию
func task1() {
	p := &Person{Name: "Bob", age: 30}
	fmt.Println(p.Name)
	changePersonPtr(p)
	fmt.Println(p.Name)
}
func changePersonPtr(p *Person) {
	p = &Person{Name: "Alice", age: 25}
}

// ЗАДАЧА 2: Что выведет?
// OUTPUT:
//
//	Bob 30
//	Alice 25   ← modifyPerson изменяет поля ЧЕРЕЗ указатель, оригинал меняется
func task2() {
	p := &Person{Name: "Bob", age: 30}
	fmt.Println(p.Name, p.age)
	modifyPerson(p)
	fmt.Println(p.Name, p.age)
}
func modifyPerson(p *Person) {
	p.Name = "Alice"
	p.age = 25
}

// ЗАДАЧА 3: Что выведет?
// OUTPUT:
//
//	{}   ← Employee имеет только неэкспортируемые поля (name, salary);
//	     json.Marshal игнорирует неэкспортируемые поля → пустой JSON объект
func task3() {
	e := Employee{name: "John", salary: 5000}
	b, _ := json.Marshal(e)
	fmt.Println(string(b))
}

// ЗАДАЧА 4: Что выведет?
// OUTPUT:
//
//	{"Value":42,"name":"test"}
//	← Data.Value имеет тег `json: "value"` (с пробелом) — НЕВАЛИДНЫЙ тег!
//	← Невалидный тег игнорируется, поле сериализуется под именем "Value" (имя поля)
//	← Data.Name имеет валидный тег `json:"name"` → "name"
func task4() {
	d := Data{Value: 42, Name: "test"}
	b, _ := json.Marshal(d)
	fmt.Println(string(b))
}

// ЗАДАЧА 5: Что выведет?
// OUTPUT:
//
//	{"host":"localhost"}   ← Config.port неэкспортирован → не сериализуется
//	{Host:localhost port:0}   ← port остаётся 0 (не десериализуется)
func task5() {
	c := Config{Host: "localhost", port: 8080}
	b, _ := json.Marshal(c)
	fmt.Println(string(b))
	var out Config
	json.Unmarshal(b, &out)
	fmt.Printf("%+v\n", out)
}

// ЗАДАЧА 6: Что выведет?
// OUTPUT:
//
//	{"ID":1,"Price":999.99}   ← Product.name неэкспортировано → не попадает в JSON
func task6() {
	p := Product{ID: 1, name: "Laptop", Price: 999.99}
	b, _ := json.Marshal(p)
	fmt.Println(string(b))
}

// ЗАДАЧА 7: Что выведет?
// OUTPUT:
//
//	Alice 30   ← p1 не изменился (struct копируется по значению)
//	Bob 25
func task7() {
	p1 := Person{Name: "Alice", age: 30}
	p2 := p1
	p2.Name = "Bob"
	p2.age = 25
	fmt.Println(p1.Name, p1.age)
	fmt.Println(p2.Name, p2.age)
}

// ЗАДАЧА 8: Что выведет?
// OUTPUT:
//
//	Bob 25   ← p2 := p1 копирует указатель; оба указывают на одну struct
func task8() {
	p1 := &Person{Name: "Alice", age: 30}
	p2 := p1
	p2.Name = "Bob"
	p2.age = 25
	fmt.Println(p1.Name, p1.age)
}

// ЗАДАЧА 9: Что выведет?
// OUTPUT:
//
//	Extended   ← e.Name обращается к полю Extended (shadowing — перекрывает Base.Name)
//	Base       ← e.Base.Name явно обращается к Base.Name
func task9() {
	e := Extended{
		Base:  Base{ID: 1, Name: "Base"},
		Name:  "Extended",
		Extra: "Data",
	}
	fmt.Println(e.Name)
	fmt.Println(e.Base.Name)
}

// ЗАДАЧА 10: Что выведет?
// OUTPUT:
//
//	Alice
//	Moscow        ← u.City продвигается из embedded Address
//	test@mail.com ← u.Email продвигается из embedded Contact
func task10() {
	u := User{
		Address: Address{City: "Moscow", Country: "Russia"},
		Contact: Contact{Email: "test@mail.com", Phone: "123"},
		Name:    "Alice",
	}
	fmt.Println(u.Name)
	fmt.Println(u.City)
	fmt.Println(u.Email)
}

// ЗАДАЧА 11: Что выведет?
// OUTPUT:
//
//	true    ← o.Inner == nil (не инициализирован)
//	test    ← o.Label доступен
//	PANIC: runtime error: invalid memory address or nil pointer dereference
//	← o.Value разыменовывает nil *Inner → panic
func task11() {
	o := Outer{Label: "test"}
	fmt.Println(o.Inner == nil)
	fmt.Println(o.Label)
	fmt.Println(o.Value) // panic: nil pointer dereference
}

// ЗАДАЧА 12: Что выведет?
// OUTPUT:
//
//	100
//	100   ← o.Value и o.Inner.Value — одно и то же поле (embedding)
func task12() {
	o := Outer{
		Inner: &Inner{Value: 42},
		Label: "test",
	}
	o.Value = 100
	fmt.Println(o.Inner.Value)
	fmt.Println(o.Value)
}

// ЗАДАЧА 13: Что выведет?
// OUTPUT:
//
//	Alice   ← range копирует значение p; изменение копии не влияет на persons[0]
func task13() {
	persons := []Person{
		{Name: "Alice", age: 30},
		{Name: "Bob", age: 25},
	}
	for _, p := range persons {
		p.Name = "Changed"
	}
	fmt.Println(persons[0].Name)
}

// ЗАДАЧА 14: Что выведет?
// OUTPUT:
//
//	Changed   ← изменение через индекс persons[i] меняет элемент напрямую
func task14() {
	persons := []Person{
		{Name: "Alice", age: 30},
		{Name: "Bob", age: 25},
	}
	for i := range persons {
		persons[i].Name = "Changed"
	}
	fmt.Println(persons[0].Name)
}

// ЗАДАЧА 15: Что выведет?
// OUTPUT:
//
//	Changed   ← slice *Person: p — указатель; p.Name = "Changed" меняет оригинал
func task15() {
	persons := []*Person{
		{Name: "Alice", age: 30},
		{Name: "Bob", age: 25},
	}
	for _, p := range persons {
		p.Name = "Changed"
	}
	fmt.Println(persons[0].Name)
}

// ЗАДАЧА 16: Что выведет?
// OUTPUT:
//
//	31   ← нельзя напрямую изменить поле map-value struct; нужно извлечь, изменить, записать
func task16() {
	m := map[string]Person{
		"alice": {Name: "Alice", age: 30},
	}
	// m["alice"].age = 31 // ошибка компиляции!
	p := m["alice"]
	p.age = 31
	m["alice"] = p
	fmt.Println(m["alice"].age)
}

// ЗАДАЧА 17: Что выведет?
// OUTPUT:
//
//	31   ← *Person в map: можно напрямую менять поля через указатель
func task17() {
	m := map[string]*Person{
		"alice": {Name: "Alice", age: 30},
	}
	m["alice"].age = 31
	fmt.Println(m["alice"].age)
}

// ЗАДАЧА 18: Что выведет?
// OUTPUT:
//
//	structs.Person{Name:"", age:0}   ← zero value struct
//	true
//	true
func task18() {
	var p Person
	fmt.Printf("%#v\n", p)
	fmt.Println(p.Name == "")
	fmt.Println(p.age == 0)
}

// ЗАДАЧА 19: Что выведет?
// OUTPUT:
//
//	true
//	true
//	*structs.Person   ← new(Person) возвращает *Person, инициализированный zero value
func task19() {
	p := new(Person)
	fmt.Println(p.Name == "")
	fmt.Println(p.age == 0)
	fmt.Printf("%T\n", p)
}

// ЗАДАЧА 20: Что выведет?
// OUTPUT:
//
//	true    ← p1 == p2: struct comparable (все поля comparable); одинаковые значения
//	false   ← p1 != p3: age разный (30 vs 31)
func task20() {
	p1 := Person{Name: "Alice", age: 30}
	p2 := Person{Name: "Alice", age: 30}
	fmt.Println(p1 == p2)
	p3 := Person{Name: "Alice", age: 31}
	fmt.Println(p1 == p3)
}

// ЗАДАЧА 21: Что выведет?
// OUTPUT:
//
//	false   ← p1 == p2 сравнивает адреса указателей (разные объекты → разные адреса)
//	true    ← *p1 == *p2 сравнивает значения struct
func task21() {
	p1 := &Person{Name: "Alice", age: 30}
	p2 := &Person{Name: "Alice", age: 30}
	fmt.Println(p1 == p2)
	fmt.Println(*p1 == *p2)
}

// ЗАДАЧА 22: Что выведет?
// OUTPUT:
//
//	5   ← value receiver: increment работает с копией; c.value не изменяется
func task22() {
	c := Counter{value: 5}
	c.increment()
	fmt.Println(c.value)
}
func (c Counter) increment() {
	c.value++
}

// ЗАДАЧА 23: Что выведет?
// OUTPUT:
//
//	6   ← pointer receiver: incrementPtr работает с оригиналом; c.value увеличивается
func task23() {
	c := Counter{value: 5}
	c.incrementPtr()
	fmt.Println(c.value)
}
func (c *Counter) incrementPtr() {
	c.value++
}

// ЗАДАЧА 24: Что выведет?
// OUTPUT:
//
//	6   ← c.incrementPtr() → Go автоматически берёт адрес: (&c).incrementPtr()
//	7   ← (&c).incrementPtr() явно
func task24() {
	c := Counter{value: 5}
	c.incrementPtr()
	fmt.Println(c.value)
	(&c).incrementPtr()
	fmt.Println(c.value)
}

// ЗАДАЧА 25: Что выведет?
// OUTPUT:
//
//	{name: salary:0}   ← json.Unmarshal не может записать в неэкспортируемые поля
func task25() {
	jsonStr := `{"name":"John","salary":5000}`
	var e Employee
	json.Unmarshal([]byte(jsonStr), &e)
	fmt.Printf("%+v\n", e)
}

// ЗАДАЧА 26: Что выведет?
// OUTPUT:
//
//	{Code:200 Message:OK data:}
//	← code=200, msg="OK" десериализуются в Code и Message
//	← data="secret" не десериализуется (поле data неэкспортировано)
//	← extra="ignored" — нет соответствующего поля (игнорируется)
func task26() {
	jsonStr := `{"code":200,"msg":"OK","data":"secret","extra":"ignored"}`
	var r Response
	json.Unmarshal([]byte(jsonStr), &r)
	fmt.Printf("%+v\n", r)
}

// ЗАДАЧА 27: Что выведет?
// OUTPUT:
//
//	{Name:Alice Age:30}
func task27() {
	p := struct {
		Name string
		Age  int
	}{
		Name: "Alice",
		Age:  30,
	}
	fmt.Printf("%+v\n", p)
}

// ЗАДАЧА 28: Что выведет?
// OUTPUT:
//
//	true   ← анонимные struct с идентичными полями comparable; значения совпадают
func task28() {
	p1 := struct {
		Name string
		Age  int
	}{"Alice", 30}
	p2 := struct {
		Name string
		Age  int
	}{"Alice", 30}
	fmt.Println(p1 == p2)
}

// ЗАДАЧА 29: Что выведет?
// OUTPUT:
//
//	{"name":"test"}    ← Value=0 → omitempty убирает; Name="test" → остаётся
//	{"value":42}       ← Name="" → omitempty убирает; Value=42 → остаётся
func task29() {
	type Data struct {
		Name  string `json:"name,omitempty"`
		Value int    `json:"value,omitempty"`
	}
	d1 := Data{Name: "test", Value: 0}
	d2 := Data{Name: "", Value: 42}
	b1, _ := json.Marshal(d1)
	b2, _ := json.Marshal(d2)
	fmt.Println(string(b1))
	fmt.Println(string(b2))
}

// ЗАДАЧА 30: Что выведет?
// OUTPUT:
//
//	{"public":"visible"}   ← тег `json:"-"` полностью исключает поле из JSON
func task30() {
	type Data struct {
		Public  string `json:"public"`
		Private string `json:"-"`
	}
	d := Data{Public: "visible", Private: "hidden"}
	b, _ := json.Marshal(d)
	fmt.Println(string(b))
}

// ЗАДАЧА 31: Что выведет?
// OUTPUT:
//
//	1
//	2
//	3   ← связный список: n1→n2→n3
func task31() {
	n1 := &Node{Value: 1}
	n2 := &Node{Value: 2}
	n3 := &Node{Value: 3}
	n1.Next = n2
	n2.Next = n3
	fmt.Println(n1.Value)
	fmt.Println(n1.Next.Value)
	fmt.Println(n1.Next.Next.Value)
}

// ЗАДАЧА 32: Что выведет?
// OUTPUT:
//
//	1
//	2
//	1   ← циклический список: n1→n2→n1; n1.Next.Next == n1
func task32() {
	n1 := &Node{Value: 1}
	n2 := &Node{Value: 2}
	n1.Next = n2
	n2.Next = n1
	fmt.Println(n1.Value)
	fmt.Println(n1.Next.Value)
	fmt.Println(n1.Next.Next.Value)
}

// ЗАДАЧА 33: Что выведет?
// OUTPUT:
//
//	20
//	20   ← b2 := b1 копирует struct по значению, но b1.Value — указатель на x
//	     ← *b2.Value = 20 меняет x через тот же указатель → b1.Value тоже = 20
func task33() {
	type Box struct {
		Value *int
	}
	x := 10
	b1 := Box{Value: &x}
	b2 := b1
	*b2.Value = 20
	fmt.Println(*b1.Value)
	fmt.Println(x)
}

// ЗАДАЧА 34: Что выведет?
// OUTPUT:
//
//	100   ← d2 := d1 копирует struct, но Values — slice (ссылочный тип)
//	      ← d2.Values[0]=100 меняет общий backing array → d1.Values[0] тоже = 100
func task34() {
	type Data struct {
		Values []int
	}
	d1 := Data{Values: []int{1, 2, 3}}
	d2 := d1
	d2.Values[0] = 100
	fmt.Println(d1.Values[0])
}

// ЗАДАЧА 35: Что выведет?
// OUTPUT:
//
//	{"id":1,"name":"test"}   ← embedded struct поля продвигаются в JSON
func task35() {
	type Base struct {
		ID int `json:"id"`
	}
	type Extended struct {
		Base
		Name string `json:"name"`
	}
	e := Extended{Base: Base{ID: 1}, Name: "test"}
	b, _ := json.Marshal(e)
	fmt.Println(string(b))
}

// ЗАДАЧА 36: Что выведет?
// OUTPUT:
//
//	{"id":1,"name":"test"}   ← *Base тоже продвигается в JSON (указатель на embedded)
func task36() {
	type Base struct {
		ID int `json:"id"`
	}
	type Extended struct {
		*Base
		Name string `json:"name"`
	}
	e := Extended{Base: &Base{ID: 1}, Name: "test"}
	b, _ := json.Marshal(e)
	fmt.Println(string(b))
}

// ЗАДАЧА 37: Что выведет?
// OUTPUT:
//
//	24   ← A{bool,int64,bool}: bool=1+7(pad)+int64=8+bool=1+7(pad) = 24 байта
//	16   ← B{bool,bool,int64}: bool=1+bool=1+6(pad)+int64=8 = 16 байт
//	← КЛЮЧЕВОЕ: порядок полей влияет на выравнивание и размер struct
func task37() {
	type A struct {
		a bool
		b int64
		c bool
	}
	type B struct {
		a bool
		c bool
		b int64
	}
	fmt.Println(unsafe.Sizeof(A{}))
	fmt.Println(unsafe.Sizeof(B{}))
}

// ЗАДАЧА 38: Что выведет?
// OUTPUT:
//
//	0   ← пустая struct имеет размер 0 байт
//	0   ← массив пустых struct тоже 0 байт! Все они указывают на одну точку памяти
func task38() {
	type Empty struct{}
	fmt.Println(unsafe.Sizeof(Empty{}))
	arr := [1000000]Empty{}
	fmt.Println(unsafe.Sizeof(arr))
}

// ЗАДАЧА 39: Что выведет?
// OUTPUT:
//
//	{"id":"123","value":"test"}   ← тег `json:"id,string"` сериализует int как JSON string
func task39() {
	type Data struct {
		ID    int    `json:"id,string"`
		Value string `json:"value"`
	}
	d := Data{ID: 123, Value: "test"}
	b, _ := json.Marshal(d)
	fmt.Println(string(b))
}

// ЗАДАЧА 40: Что выведет?
// OUTPUT:
//
//	true   ← json.Unmarshal возвращает ошибку (string в int поле)
//	0      ← d.Value не изменился (остался zero value)
func task40() {
	type Data struct {
		Value int `json:"value"`
	}
	jsonStr := `{"value":"not a number"}`
	var d Data
	err := json.Unmarshal([]byte(jsonStr), &d)
	fmt.Println(err != nil)
	fmt.Println(d.Value)
}

// ЗАДАЧА 41: Что выведет?
// OUTPUT:
//
//	1 2
//	← c.Value — ambiguous selector (A.Value и B.Value); нужно явно указывать c.A.Value
func task41() {
	type A struct {
		Value int
	}
	type B struct {
		Value int
	}
	type C struct {
		A
		B
	}
	c := C{A: A{Value: 1}, B: B{Value: 2}}
	// fmt.Println(c.Value) // ошибка: ambiguous selector
	fmt.Println(c.A.Value, c.B.Value)
}

// ЗАДАЧА 42: Что выведет?
// OUTPUT:
//
//	11   ← s.increase() вызывает (m *Metrics).increase() через embedding; Count: 10→11
//	11   ← s.Count и s.Metrics.Count — одно и то же поле
func task42() {
	s := Stats{
		Metrics: Metrics{Count: 10},
		Total:   100,
	}
	s.increase()
	fmt.Println(s.Count)
	fmt.Println(s.Metrics.Count)
}
func (m *Metrics) increase() {
	m.Count++
}

// ЗАДАЧА 43: Что выведет?
// OUTPUT:
//
//	Alice   ← changePtr(**Person) получает **Person; *pp = ... меняет оригинальный *Person
func task43() {
	p := &Person{Name: "Bob", age: 30}
	changePtr(&p)
	fmt.Println(p.Name)
}
func changePtr(pp **Person) {
	*pp = &Person{Name: "Alice", age: 25}
}

// ЗАДАЧА 44: Что выведет?
// OUTPUT:
//
//	10 20
//	← Point{20} — ошибка компиляции (позиционная инициализация требует все поля)
func task44() {
	p := Point{10, 20}
	fmt.Println(p.X, p.Y)
	//q := Point{20} что тут будет?
	//fmt.Println(q)
}

// ЗАДАЧА 45: Что выведет?
// OUTPUT:
//
//	10 0   ← Point{X:10}: Y не указан → zero value = 0
func task45() {
	p := Point{X: 10}
	fmt.Println(p.X, p.Y)
}

// ЗАДАЧА 46: Что выведет?
// OUTPUT:
//
//	test
//	{"nested":"value"}   ← json.RawMessage хранит raw JSON без парсинга
func task46() {
	type Data struct {
		Name string          `json:"name"`
		Raw  json.RawMessage `json:"raw"`
	}
	jsonStr := `{"name":"test","raw":{"nested":"value"}}`
	var d Data
	json.Unmarshal([]byte(jsonStr), &d)
	fmt.Println(d.Name)
	fmt.Println(string(d.Raw))
}

// ЗАДАЧА 47: Что выведет?
// OUTPUT:
//
//	{"status":200,"data":{"name":"test","value":42}}   ← вложенный анонимный struct в JSON
func task47() {
	type Response struct {
		Status int `json:"status"`
		Data   struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"data"`
	}
	r := Response{Status: 200}
	r.Data.Name = "test"
	r.Data.Value = 42
	b, _ := json.Marshal(r)
	fmt.Println(string(b))
}

// ЗАДАЧА 48: Что выведет?
// OUTPUT:
//
//	{"Value":42,"name":"test"}
//	← `json: "value"` (с пробелом) — невалидный тег → используется имя поля "Value"
//	← `json:"name"` — валидный тег → "name"
func task48() {
	type Data struct {
		Value int    `json: "value" xml:"val"`
		Name  string `json:"name"`
	}
	d := Data{Value: 42, Name: "test"}
	b, _ := json.Marshal(d)
	fmt.Println(string(b))
}

// ЗАДАЧА 49: Что выведет?
// OUTPUT:
//
//	{Data:{Value:0} Name:}   ← zero value Outer: Inner.Value=0, Name=""
//	0
func task49() {
	type Inner struct {
		Value int
	}
	type Outer struct {
		Data Inner
		Name string
	}
	var o Outer
	fmt.Printf("%+v\n", o)
	fmt.Println(o.Data.Value)
}

// ЗАДАЧА 50: Что выведет?
// OUTPUT:
//
//	{Name:Alice Age:30}
//	← PersonA и PersonB структурно идентичны, но разные типы
//	← прямое присваивание b := PersonB(a) работает (explicit conversion между struct с одинаковыми полями)
//	← ИСПРАВЛЕНИЕ: закомментированный код работает, а не является ошибкой!
func task50() {
	type PersonA struct {
		Name string
		Age  int
	}
	type PersonB struct {
		Name string
		Age  int
	}
	a := PersonA{Name: "Alice", Age: 30}
	// b := PersonB(a) // ошибка компиляции!
	b := PersonB{Name: a.Name, Age: a.Age}
	fmt.Printf("%+v\n", b)
}
