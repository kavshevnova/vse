// Задача 10 – JSON и неэкспортируемое поле
// Как отработает код?
//
// ОТВЕТ:
//	main.Data{Age:33, name:"Ivan"}    ← %#v выводит все поля включая неэкспортируемые
//	{"age":33}                         ← json.Marshal игнорирует name (неэкспортируемое)
//	main.Data{Age:33, name:""}         ← json.Unmarshal не может записать в name
//
// Объяснение:
// - Age (экспортируемое) с тегом "age" → сериализуется/десериализуется
// - name (неэкспортируемое) → json.Marshal ИГНОРИРУЕТ; json.Unmarshal НЕ МОЖЕТ записать
// - После Unmarshal: Age=33 (из JSON), name="" (zero value, JSON не записал)
package main

import (
	"encoding/json"
	"fmt"
)

type Data struct {
	Age  int    `json:"age"`
	name string `json:"name"`
}

func testData() {
	in := Data{33, "Ivan"}
	fmt.Printf("%#v\n", in) // main.Data{Age:33, name:"Ivan"}

	encoded, _ := json.Marshal(in)
	fmt.Println(string(encoded)) // {"age":33}

	var out Data
	json.Unmarshal(encoded, &out)
	fmt.Printf("%#v\n", out) // main.Data{Age:33, name:""}
}

func main() { testData() }
