```go
package main

import (
	"fmt"
	"os"
)

func Foo() error {
	var err *os.PathError = nil
	return err
}

func main() {
	err := Foo()
	fmt.Println(err)        // <nil>
	fmt.Println(err == nil) // false
	//вывод: <nil> false
}
```
Интерфейс в Го можно представить как
структуру с полями type и value (под капотом все устроенно чуть сложнее)
и когда мы проверяем интерфейс на nil мы проверяем значение
обоих полей (и type и value), но значение type!=nil, поэтому мы получим вывод:
вывод:
\<nil>
false

