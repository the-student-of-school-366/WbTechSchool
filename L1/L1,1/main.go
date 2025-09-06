package main

import (
	"fmt"
)

// Human базовая структура
type Human struct {
	name    string
	age     int
	surname string
}

// SayHello возвращает строку с именем и фамилией Human
func (h Human) SayHello() string {
	return fmt.Sprintf("Hello, my name is %s %s!\n", h.name, h.surname)
}

// Action имеет доступ ко всем полям и методам Human
type Action struct {
	Human
	str string
}

func main() {

	acy := Action{
		Human: Human{
			name:    "Ananda",
			age:     20,
			surname: "Uldanov",
		},

		str: "act",
	}
	//вызываем метод напрямую
	fmt.Println("Acy act:", acy.SayHello())

	//прописываем полный путь для метода
	fmt.Println("Also acy str:", acy.Human.SayHello())

}
