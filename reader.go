// reader
package main

import (
	"encoding/hex"
	"fmt"
	"io"
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

/*
    Create a new reader on device/port. Baud rate defaults to 9600
	if too low [<110]. The read queue is maxed to 100K bytes.
*/
func NewReader(device string, speed int) *Reader {
	r := new(Reader)
	r.device = device
	r.speed = uint(speed)
	if r.speed < 110 {
		r.speed = 9600
	}
	// Use a queue for 100K bytes
	r.dataqueue = NewQueue(100000)
	r.lastUpdate = time.Now()

	return r
}

/*
	Get a pointer to the queue for reading
*/
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
			status := r.initLogger(false)
			if !status {
				r.dataqueue.Clear()
				time.Sleep(10 * time.Minute)
			}
		}
	}
}

/*
	Opens (and closes) the communication port and initializes the Growatt
	inverter to start sending the datagram	data. It *should* only send
	every 1.5 seconds, but currently I receive data continuously.
*/
func (r *Reader) initLogger(silent bool) bool {
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

	var status bool

	status = r.sendCommand(conn, "Init", []byte{
		0x3F, 0x23, 0x7E, 0x34, 0x41, 0x7E, 0x32,
		0x59, 0x31, 0x35, 0x30, 0x30, 0x23, 0x3F})

	if !status {
		return false
	}

	status = r.sendCommand(conn, "Commit", []byte{
		0x3F, 0x23, 0x7E, 0x34, 0x42, 0x7E, 0x23, 0x3F})

	Info("Sent init command to Growatt inverter.")
	return status
}

func (r *Reader) sendCommand(conn io.ReadWriteCloser, task string, data []byte) bool {
	_, err1 := conn.Write(data)
	if err1 != nil {
		log.Fatalf("[ERROR] serial.sendCommand: %v", err1)
	}
	time.Sleep(250 * time.Millisecond)

	// Read the arbitrarily data until InterCharacterTimeout
	buffer := make([]byte, 64)
	size, err2 := conn.Read(buffer)
	if err2 != nil {
		Warn(task + " not accepted: " + err2.Error())
		return false
	}
	if size == 0 {
		Warn(task + " not accepted: Empty response.")
		return false
	}

	// If all fields are ff the system isn't started yet.
	// If all fields are de the system is shutting down.
	equal := true
	first := buffer[0]
	for i := 1; i < size; i++ {
		if buffer[i] != first {
			equal = false
			break
		}
	}

	if equal {
		Warn(task + " not accepted: Code " + string(first) + ".")
		return false
	}

	Verbose("Reading size of send command: " + strconv.Itoa(size))
	Verbose(hex.Dump(buffer[0:size]))
	return true
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

	buffer := make([]byte, 30)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Warn("Reading failed due to: " + err.Error())
			break
		}

		span := time.Now().Sub(r.lastUpdate)
		r.lastUpdate = time.Now()
		if span > 5*time.Minute {
			fmt.Println()
			Warn("Respawning...")
			r.dataqueue.Clear()
			return true
		}

		if !reading {
			reading = true
			Info("Reading started with " + strconv.Itoa(n) + " bytes.")
		}

		// TODO Error because it keeps on reading and getting data. How to stop it?
		// Verbose("Read bytes and pushing: " + strconv.Itoa(n))

		for i := 0; i < n; i++ {
			r.dataqueue.Push(buffer[i])
		}
	}
	Warn("Reading stopped.")
	return false
}
