package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"growattrr/diag"
	"growattrr/reader"
)

type Interpreter struct {
	inputQueue *reader.Queue
	lastData   *Datagram
	lock       *sync.Mutex
	hasSlept   bool
	sleeping   bool
	status     string
	lastUpdate time.Time
}

type Datagram struct {
	Power           float32
	VoltagePV1      float32
	VoltagePV2      float32
	VoltageBus      float32 `json:",omitempty"`
	VoltageGrid     float32 `json:",omitempty"`
	TotalProduction float32 `json:",omitempty"`
	DayProduction   float32
	Frequency       float32 `json:",omitempty"`
	Temperature     float32 `json:",omitempty"`
	OperationHours  float32 `json:",omitempty"`
	Status          string
	FaultCode       int
	Timestamp       time.Time
}

func HomeAssistantConfig() [13][4]string {
	return [...][4]string{
		{"Power", "power", "W", "6dbbd634-cfcc-4ecf-b2e8-130708511b24"},
		{"VoltagePV1", "voltage", "V", "fc65022e-2db6-405d-9900-80f981b42c21"},
		{"VoltagePV2", "voltage", "V", "6354256e-520f-48d7-a2da-dc307fcfddf5"},
		{"VoltageBus", "voltage", "V", "a95a0291-4340-4b93-a07b-aa1b17e917a8"},
		{"VoltageGrid", "voltage", "V", "8ab07337-e58a-473b-b61b-560596c9621c"},
		{"TotalProduction", "energy", "kWh", "496138a6-aba6-4aca-a47d-9b9aa7aa05a4"},
		{"DayProduction", "energy", "kWh", "14ef721f-8225-44e1-bbc9-ebe603ea1d81"},
		{"Frequency", "frequency", "Hz", "851bec08-3410-4227-8e84-5cbcb176329a"},
		{"Temperature", "temperature", "°C", "490abec4-b9a0-4b34-933b-a1206fa51cc0"},
		{"OperationHours", "duration", "h", "5d82c4fd-8fb6-4465-a402-fcde57ac464f"},
		{"Status", "", "", "edcd66f5-dc8a-442b-b1f1-fd599dbbf4b0"},
		{"FaultCode", "", "", "5695793e-8dfb-4a8e-89d6-7bfb9cf99cd8"},
		{"Timestamp", "date", "", "21b5c51c-2e87-4b57-a999-02375e033bf2"},
	}
}

func NewInterpreter(inque *reader.Queue) *Interpreter {
	i := new(Interpreter)
	i.inputQueue = inque
	i.lock = &sync.Mutex{}
	i.hasSlept = false
	return i
}

func NewDatagram() *Datagram {
	dg := new(Datagram)
	dg.Timestamp = time.Now().Round(time.Second)
	dg.Status = "Unavailable"
	dg.FaultCode = -1
	return dg
}

/*
	Reads from the input queue until a valid size block has been read
	for creating a datagram. It will go into sleep mode if no data is
	received within several seconds.
*/
func (i *Interpreter) start() {
	diag.Info("Start interpreter...")
	buffer := make([]byte, 40, 40)
	idx := 0
	errCount := 0
	emptyCount := 0

	for {
		e := i.inputQueue.Pop()
		if e == nil {
			i.status = "Polling since " + i.lastUpdate.Format("15:04:05")
			// No input yet. Let's sleep...
			time.Sleep(100 * time.Millisecond)

			// If empty for long (10 seconds), clear data and lock for 5 min.
			emptyCount = emptyCount + 1
			if emptyCount == 100 {
				emptyCount = 0
				i.status = "Not receiving"
				if !i.hasSlept {
					diag.Warn("Processing will sleep now.")
				}
				i.updateToDatagram("Sleeping")

				// Sleep
				i.inputQueue.Clear()
				i.sleeping = true
				time.Sleep(5 * time.Minute)
				i.hasSlept = true
			}
		} else {
			i.status = "Receiving data"
			emptyCount = 0
			i.lastUpdate = time.Now()

			if i.hasSlept {
				diag.Info("Processing will proceed!")
				i.hasSlept = false
				i.sleeping = false
			}
			bv := e.(byte)
			if bv == 0x57 && idx >= 30 {
				i.status = "Supplying datagrams"
				i.createAndStoreDatagram(buffer[0:idx])
				idx = 0
			} else if idx >= 40 {
				i.status = "Receiving wrong data"
				i.updateToDatagram("Invalid")

				diag.Verbose(hex.Dump(buffer[0:idx]))

				idx = 0
				errCount = errCount + 1
				if errCount > 20 {
					diag.Warn("Invalid data received. Waiting...")
					i.status = "Awaiting correct data"
					time.Sleep(5 * time.Minute)
					i.inputQueue.Clear()
					errCount = 0
				} else {
					diag.Warn("Invalid data received. Retrying...")
				}
			} else {
				i.status = "Reading"
				buffer[idx] = bv
				idx = idx + 1
			}
		}
	}
}

func (i *Interpreter) updateToDatagram(status string) {
	// Update to an empty datagram with updated time
	i.lock.Lock()
	// But keep the accumulated data
	var day float32
	var total float32
	if i.lastData != nil {
		day = i.lastData.DayProduction
		total = i.lastData.TotalProduction
	}
	i.lastData = NewDatagram()
	i.lastData.TotalProduction = total
	i.lastData.DayProduction = 0
	i.lastData.Status = status

	if i.lastUpdate.Day() == time.Now().Day() {
		i.lastData.DayProduction = day
	}
	i.lock.Unlock()
}

/*
	Processes the 30 bytes to a valid datagram. If less or more bytes are
	given, an error is produced and the data is dumped on screen.
*/
func (i *Interpreter) createAndStoreDatagram(data []byte) {
	if len(data) != 30 {
		diag.Warn("Datagram incorrect size; ignoring " + strconv.Itoa(len(data)) + " bytes ...")
		diag.Verbose(hex.Dump(data))
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
		// Occurs at start and end of active period.
		dg.Status = "Waiting"
	case 1:
		dg.Status = "Normal"
	case 2:
		dg.Status = "Fault"
	}

	i.lock.Lock()
	i.lastData = dg
	i.lock.Unlock()
}

/* Retrieves the latest datagram as interpreted. */
func (i *Interpreter) getDatagram() *Datagram {
	i.lock.Lock()
	result := i.lastData
	i.lock.Unlock()
	return result
}

/* Convert single byte to integer */
func (i *Interpreter) decodeSmallValue(data byte) int {
	return int(data)
}

/* Convert double byte to integer */
func (i *Interpreter) decodeValue(msw byte, lsw byte, div int) float32 {
	return float32(int(msw)*256+int(lsw)) / float32(div)
}

/* Convert quad byte to integer */
func (i *Interpreter) decodeLargeValue(data []byte, div int) float32 {
	return float32(i.decodeValue(data[0], data[1], 1)*65536+
		i.decodeValue(data[2], data[3], 1)) / float32(div)
}

/* Datagram to string function */
func (d Datagram) String() string {
	result, _ := json.Marshal(d)
	return fmt.Sprintf("[Datagram] %v", string(result))
}
