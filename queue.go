package main

import "sync"

type Element struct {
	data interface{}
}

type Queue struct {
	lock      *sync.Mutex // Parameter to add call synchronisation
	container []Element   // Holder of all elements in the queue
	maxSize   int         // Requested max size
	counter   int         // Used for counting the size
}

func NewQueue(size int) *Queue {
	qd := new(Queue)
	qd.lock = &sync.Mutex{}
	qd.maxSize = size
	if size < 100 {
		qd.maxSize = 100
	}
	qd.counter = 0

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
	qd.counter = qd.counter + 1
	
	// Shrink if needed with a third of the current size
	if qd.counter > qd.maxSize {
		shrinkSize := len(qd.container) / 3
		qd.container = qd.container[shrinkSize:]
		qd.counter = len(qd.container)
	}
	qd.lock.Unlock()
}

/*
	Clear the queue.
*/
func (qd *Queue) Clear() {
	qd.lock.Lock()
	qd.container = make([]Element, 0)
	qd.counter = 0
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
	qd.counter = qd.counter - 1
	qd.lock.Unlock()

	return r
}
