package main

import "fmt"

type ITransport interface {
	Move()
}

type Car struct {
	MoveSpeed int
}

func (car *Car) Move() {
	fmt.Printf("Moooove with speed: %d\n", car.MoveSpeed)
}

type IPerson interface {
	Walk()
}

type Man struct {
	WalkSpeed int
}

func (man *Man) Walk() {
	fmt.Printf("Walking with speed: %d\n", man.WalkSpeed)
}

type Adapter struct {
	MyMan IPerson
}

func NewAdapter(myMan IPerson) *Adapter {
	return &Adapter{MyMan: myMan}
}

func (adapter *Adapter) Move() {
	adapter.MyMan.Walk()
}

func UseOnlyMove(transport ITransport) {
	transport.Move()
}

func main() {
	myMan := Man{20}
	adapter := NewAdapter(&myMan)
	UseOnlyMove(adapter)
}
