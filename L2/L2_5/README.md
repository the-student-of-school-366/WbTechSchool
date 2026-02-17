```go
package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return nil
}

func main() {
	var err error // err - интерфейс error (имеет тип и значение)
	err = test()
	if err != nil { // err не nil, так как интерфейс имеет тип error
		println("error")
		return
	}
	println("ok")
}
//Вывод: error

```
Переменная err объявляется как интерфейс error, который имеет тип (*customError)
и значение (nil). Интерфейс равен nil только тогда когда и тип и значение nil.
Выведется error