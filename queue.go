package main

import "sync"

type Element struct {
	data interface{}
}

type Queue struct {
	lock      *sync.Mutex		// Parameter to add call synchronisation
	container []Element			// Holder of all elements in the queue
}

func NewQueue() *Queue {
	qd := new(Queue)
	qd.lock = &sync.Mutex{}

	return qd
}

/*
	Push any element on the queue. Synchronized add to end.
*/
func (qd *Queue) Push(data interface{}) {
	qd.lock.Lock()

	element := new(Element)
	element.data = data
	qd.container = append(qd.container, *element)

	qd.lock.Unlock()
}

/*
	Clear the queue.
*/
func (qd *Queue) Clear() {
	qd.lock.Lock()
	qd.container = make([]Element, 0)
	qd.lock.Unlock()
}

/*
	Pop the first added element off the queue. Nil if empty.
	Note this is FIFO (first in first out) behavior. 
*/
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
