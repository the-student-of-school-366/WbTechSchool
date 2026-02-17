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
