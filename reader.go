// reader
package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type Reader struct {
	connection io.ReadWriteCloser
	dataqueue  *Queue
	device     string
	speed      uint
	lastUpdate time.Time
	started    bool
}

func NewReader(device string, speed int) *Reader {
	r := new(Reader)
	r.device = device
	r.speed = uint(speed)
	if r.speed == 0 {
		r.speed = 9600
	}
	r.dataqueue = NewQueue()
	r.lastUpdate = time.Now()

	return r
}

func (r *Reader) getQueue() *Queue {
	return r.dataqueue
}

func (r *Reader) sendInitCommand(conn io.ReadWriteCloser, silent bool) {
	initString1 := []byte{0x3F, 0x23, 0x7E, 0x34, 0x41, 0x7E, 0x32, 0x59, 0x35, 0x30, 0x30, 0x30, 0x23, 0x3F}
	initString2 := []byte{0x3F, 0x23, 0x7E, 0x34, 0x42, 0x7E, 0x23, 0x3F}

	_, err1 := conn.Write(initString1)
	if err1 != nil {
		log.Fatalf("[ERROR] serial.sendInitCommand [1]: %v", err1)
	}

	// Read the arbitrarily data until InterCharacterTimeout
	buffer := make([]byte, 256)
	_, err2 := conn.Read(buffer)
	if err2 != nil {
		log.Fatalf("[ERROR] serial.sendInitCommand [2]: %v", err2)
	}

	_, err3 := conn.Write(initString2)
	if err3 != nil {
		log.Fatalf("[ERROR] serial.sendInitCommand [3]: %v", err3)
	}

	//	if !silent {
	Info("Sent init command to Growatt inverter.")
	//	}

	time.Sleep(250 * time.Millisecond)
}

func (r *Reader) initLogger(silent bool) {
	options := serial.OpenOptions{
		PortName:              r.device,
		BaudRate:              r.speed,
		DataBits:              8,
		StopBits:              1,
		ParityMode:            0,
		MinimumReadSize:       30,
		InterCharacterTimeout: 20,
		RTSCTSFlowControl:     false,
	}
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	defer conn.Close()
	r.sendInitCommand(conn, silent)
}

// func (r *Reader) monitorConnection() {

// 	for r.started == true {
// 		span := time.Now().Sub(r.lastUpdate)

// 		if span > 5*time.Minute {
// 			r.lastUpdate = time.Now()
// 			Warn("Closing overdue connection and restarting...")
// 			r.connection.Close()
// 			r.initLogger(true)
// 			r.start(true)
// 		}
// 		time.Sleep(1 * time.Minute)
// 	}
// 	Warn("Reader stopped...")
// }

func (r *Reader) startMonitored() {
	for {
		Info("Serial reader starting.")
		r.start(false)
	}
}

func (r *Reader) start(silent bool) {

	r.started = true

	options := serial.OpenOptions{
		PortName:        r.device,
		BaudRate:        r.speed,
		DataBits:        8,
		StopBits:        1,
		ParityMode:      0,
		MinimumReadSize: 30,

		RTSCTSFlowControl: false,
	}

	if !silent {
		Info(fmt.Sprintf("Connecting to %v [%v,8,N,1]", r.device, r.speed))
	}

	// Open the port.
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	r.connection = conn

	// Make sure to close it later.
	defer r.connection.Close()

	// go r.monitorConnection()

	buffer := make([]byte, 16)
	zeroCounter := 0
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Warn("Reading failed due to: " + err.Error())
			break
		}

		fmt.Print(".")

		span := time.Now().Sub(r.lastUpdate)
		if span > 5*time.Minute {
			Warn("Respawning...")
			r.sendInitCommand(conn, false)
		}
		r.lastUpdate = time.Now()

		for i := 0; i < n; i++ {
			if buffer[i] == 0x00 {
				zeroCounter = zeroCounter + 1
				if zeroCounter > 200 {
					Warn("Reinitialising connection (and cleaning buffer)...")
					r.dataqueue.Clear()
					r.sendInitCommand(conn, false)
				}
			} else {
				zeroCounter = 0
			}
			r.dataqueue.Push(buffer[i])
		}
	}
	Warn("Reading stopped.")
	r.started = false
}
