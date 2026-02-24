// Задача 8 – error и nil-интерфейсы
// Как отработает код?
//
// ОТВЕТ:
//	true    ← e1: var e1 error (interface nil) → nil == nil → true
//	false   ← e2: var e2 *errorString → тип *errorString + значение nil → НЕ nil interface!
//	false   ← e2 = &errorString{} → тип *errorString + значение non-nil → точно не nil
//	false   ← e3 = (*errorString)(nil) → тот же случай: тип задан, значение nil → не nil interface
//
// КЛЮЧЕВОЕ: interface == nil только если и тип, и значение == nil.
// При передаче (*errorString)(nil) в error-интерфейс создаётся interface с типом *errorString и nil-значением.
// Такой interface != nil!
//
// Это классический Go gotcha с nil interfaces.
package main

import "fmt"

type errorString struct {
	s string
}

func (e errorString) Error() string {
	return e.s
}

func checkErr(err error) {
	fmt.Println(err == nil)
}

func main() {
	var e1 error // interface nil
	checkErr(e1) // true

	var e2 *errorString // (*errorString)(nil)
	checkErr(e2)        // false — тип задан!

	e2 = &errorString{} // non-nil pointer
	checkErr(e2)        // false

	var e3 *errorString = nil // то же самое
	checkErr(e3)              // false
}
