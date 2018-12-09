package main

import (
	"encoding/hex"
	"fmt"
	"time"
)

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
	fmt.Println("Start interpreter...")
	//	var buffer []byte
	buffer := make([]byte, 40, 40)
	idx := 0
	errCount := 0

	for {
		e := i.inputQueue.Pop()
		if e == nil {
			// No input yet. Let's sleep...
			time.Sleep(100 * time.Millisecond)
		} else {
			bv := e.(Element).data.(byte)
			if bv == 0x57 {
				fmt.Println("Found datagram")
				fmt.Println("[INFO] Bytes: " + hex.Dump(buffer[0:idx]))
				i.createAndStoreDatagram(buffer, idx+1)
				idx = 0
			} else if idx >= 40 {
				fmt.Println("[WARN] Invalid data received. Retrying...")
				fmt.Println("[INFO] Bytes: " + hex.Dump(buffer))
				idx = 0
				errCount = errCount + 1
				if errCount > 20 {
					fmt.Println("[WARN] Sleeping for a while...")
					time.Sleep(30 * time.Second)
				}
			} else {
				buffer[idx] = bv
				idx = idx + 1
			}
		}
	}
}

func (i *Interpreter) createAndStoreDatagram(data []byte, size int) *Datagram {
	fmt.Println("Storing datagram with %v", data)
	return nil
}

func (i *Interpreter) pop() *Datagram {
	datagram := i.outputQueue.Pop()
	if datagram != nil {
		return datagram.(*Datagram)
	}
	return nil
}

func (d Datagram) String() string {
	return fmt.Sprintf("[Datagram] total:%v, day:%v", d.totalProduction, d.dayProduction)
}
