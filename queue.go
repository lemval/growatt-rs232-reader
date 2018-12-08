package main

import "sync"

type Element struct {
	data interface{}
}

type Queue struct {
	lock      *sync.Mutex
	container []Element
}

func (qd *Queue) Push(data interface{}) {
	qd.lock.Lock()

	element := new(Element)
	element.data = data
	qd.container = append(qd.container, *element)

	qd.lock.Unlock()
}

func (qd *Queue) Pop() interface{} {
	if len(qd.container) == 0 {
		return nil
	}
	qd.lock.Lock()
	r := qd.container[0]
	qd.container = qd.container[1:]

	qd.lock.Unlock()

	return r
}

func NewQueue() *Queue {
	qd := new(Queue)
	qd.lock = &sync.Mutex{}

	return qd
}
