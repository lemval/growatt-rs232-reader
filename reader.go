// reader
package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type Reader struct {
	dataqueue  *Queue
	device     string
	speed      uint
	lastUpdate time.Time
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

/*
	Starts and monitors the serial reader. If it terminates, it will restart
	the reader (with possible reinitialisation of the inverter on wakeup).
*/
func (r *Reader) startMonitored() {
	for {
		Info("Serial reader starting.")
		if r.start() {
			r.initLogger(false)
		}
	}
}

/*
	Opens (and closes) the communication port and initializes the Growatt
	inverter to start sending the datagram	data. It *should* only send
	every 1.5 seconds, but currently I receive data continuously.
*/
func (r *Reader) initLogger(silent bool) {
	Info("Inverter about to be initialed...")
	options := serial.OpenOptions{
		PortName:              r.device,
		BaudRate:              r.speed,
		DataBits:              8,
		StopBits:              1,
		ParityMode:            0,
		InterCharacterTimeout: 500,
		RTSCTSFlowControl:     false,
	}
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	defer conn.Close()

	initString1 := []byte{0x3F, 0x23, 0x7E, 0x34, 0x41, 0x7E, 0x32, 0x59, 0x31, 0x35, 0x30, 0x30, 0x23, 0x3F}
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

	time.Sleep(250 * time.Millisecond)
	Info("Sent init command to Growatt inverter.")
}

/*
	Starts reading until read failure or respawn of the inverter
*/
func (r *Reader) start() bool {
	options := serial.OpenOptions{
		PortName:          r.device,
		BaudRate:          r.speed,
		DataBits:          8,
		StopBits:          1,
		ParityMode:        0,
		MinimumReadSize:   30,
		RTSCTSFlowControl: false,
	}

	Info(fmt.Sprintf("Connecting to %v [%v,8,N,1]", r.device, r.speed))

	// Open the port.
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	// Make sure to close it later.
	defer conn.Close()

	reading := false

	buffer := make([]byte, 512)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Warn("Reading failed due to: " + err.Error())
			break
		}

		span := time.Now().Sub(r.lastUpdate)
		if span > 5*time.Minute {
			Warn("Respawning...")
			r.dataqueue.Clear()
			return true
		}
		r.lastUpdate = time.Now()

		if !reading {
			reading = true
			Info("Reading started with " + strconv.Itoa(n) + " bytes.")
		}
		for i := 0; i < n; i++ {
			r.dataqueue.Push(buffer[i])
		}
	}
	Warn("Reading stopped.")
	return false
}
