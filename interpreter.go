package main

import "fmt"

type Interpreter struct {
	inputQueue  *Queue
	outputQueue *Queue
}

type Datagram struct {
	totalProduction int64
	dayProduction   int32
}

func NewInterpreter(inque *Queue) *Interpreter {
	i := new(Interpreter)
	i.inputQueue = inque
	i.outputQueue = NewQueue()
	return i
}

func (i *Interpreter) start() {
	fmt.Println("Got InputQueue: %+v", i.inputQueue)
}

func (i *Interpreter) pop() *Datagram {
	fmt.Println("Got OutputQueue: %+v", i.outputQueue)
	datagram := i.outputQueue.Pop()
	if datagram != nil {
		return datagram.(*Datagram)
	}
	return nil
}

func (d Datagram) String() string {
	return fmt.Sprintf("[Datagram] total:%v, day:%v", d.totalProduction, d.dayProduction)
}
