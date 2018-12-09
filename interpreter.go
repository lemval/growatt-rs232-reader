package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Interpreter struct {
	inputQueue  *Queue
	outputQueue *Queue
}

type Datagram struct {
	VoltagePV1      float32
	VoltagePV2      float32
	VoltageBus      float32
	VoltageGrid     float32
	TotalProduction float32
	DayProduction   float32
	Frequency       float32
	Power           float32
	Temperature     float32
	OperationHours  float32
	Status          string
	FaultCode       int
	Timestamp       time.Time
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
			if bv == 0x57 && idx >= 30 {
				// TODO 0x57 might be part of the values too ...

				// fmt.Println("Found datagram")
				// fmt.Println(hex.Dump(buffer[0:idx]))
				i.createAndStoreDatagram(buffer[0:idx])
				idx = 0
			} else if idx >= 40 {
				fmt.Println("[WARN] Invalid data received. Retrying...")
				fmt.Println(hex.Dump(buffer))
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

func (i *Interpreter) createAndStoreDatagram(data []byte) {
	if len(data) != 30 {
		// 11, 18, 29
		fmt.Println("Datagram incorrect size; ignoring", len(data), "bytes ...")
		fmt.Println(hex.Dump(data))
		return
	}

	dg := new(Datagram)
	dg.VoltagePV1 = i.decodeValue(data[0], data[1], 10)
	dg.VoltageBus = i.decodeValue(data[2], data[3], 10)
	dg.VoltagePV2 = i.decodeValue(data[4], data[5], 10)
	dg.VoltageGrid = i.decodeValue(data[6], data[7], 10)
	dg.Frequency = i.decodeValue(data[8], data[9], 100)
	dg.Power = i.decodeValue(data[10], data[11], 10)
	dg.Temperature = i.decodeValue(data[12], data[13], 10)
	dg.DayProduction = i.decodeValue(data[20], data[21], 10)
	dg.TotalProduction = i.decodeLargeValue(data[22:26], 10)
	dg.OperationHours = i.decodeLargeValue(data[26:30], 7200)
	dg.FaultCode = i.decodeSmallValue(data[15])
	dg.Timestamp = time.Now()

	status := i.decodeSmallValue(data[14])
	switch status {
	case 0:
		dg.Status = "Waiting"
	case 1:
		dg.Status = "Normal"
	case 2:
		dg.Status = "Fault"
	}

	i.outputQueue.Push(dg)
}

func (i *Interpreter) decodeSmallValue(data byte) int {
	return int(data)
}

func (i *Interpreter) decodeValue(msw byte, lsw byte, div int) float32 {
	return float32((int(msw)*256 + int(lsw))) / float32(div)
}

func (i *Interpreter) decodeLargeValue(data []byte, div int) float32 {
	return float32(i.decodeValue(data[0], data[1], 1)*65536+
		i.decodeValue(data[2], data[3], 1)) / float32(div)
}

func (i *Interpreter) pop() *Datagram {
	e := i.outputQueue.Pop()
	if e != nil {
		return e.(Element).data.(*Datagram)
	}
	return nil
}

func (d Datagram) String() string {
	result, _ := json.Marshal(d)
	return fmt.Sprintf("[Datagram] %v", string(result))
}
