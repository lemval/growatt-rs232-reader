// reader
package reader

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"growattrr/diag"

	"github.com/jacobsa/go-serial/serial"
)

type Reader struct {
	dataqueue  *Queue
	device     string
	speed      uint
	lastUpdate time.Time
	Status     string
	InitStatus string
	connection io.ReadWriteCloser
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
	r.Status = "Created"
	r.InitStatus = "None"
	return r
}

/*
	Get a pointer to the queue for reading
*/
func (r *Reader) GetQueue() *Queue {
	return r.dataqueue
}

/*
	Starts and monitors the serial reader. If it terminates, it will restart
	the reader (with reinitialisation of the inverter on wakeup). If init
	fails, queue will be cleared and sleep of 10 minutes is induced.
*/
func (r *Reader) StartMonitored() {

	go r.startPoking()

	for {
		if strings.Compare(r.InitStatus, "OK") != 0 {
			r.dataqueue.Clear()
			r.InitLogger()
		}

		// count := 0
		diag.Info("Serial reader starting.")
		status := r.start()
		r.Status = "Stopped reading."
		diag.Warn("Reading stopped (" + strconv.FormatBool(status) + ").")
	}
}

/*
	Checks if the reader needs to be triggered since conn.Read is blocking
*/
func (r *Reader) startPoking() {
	for {
		time.Sleep(1 * time.Minute)
		span := time.Now().Sub(r.lastUpdate)
		if span > 5*time.Minute {
			diag.Warn("Poke needed. Reader can't read data.")
			_ = r.connection.Close()
			r.InitLogger()
			diag.Verbose("Poke done.")
		} else {
			diag.Verbose("No poke needed.")
		}
	}
}

/*
	Opens (and closes) the communication port and initializes the Growatt
	inverter to start sending the datagram	data. It *should* only send
	every 1.5 seconds, but currently I receive data continuously.
*/
func (r *Reader) InitLogger() bool {
	diag.Info("Sending initialisation to inverter...")
	r.InitStatus = "Starting"
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
		r.InitStatus = "Failed to open connection"
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	defer conn.Close()
	r.InitStatus = "Initializing"

	status := r.sendCommand(conn, "Init", []byte{
		0x3F, 0x23, 0x7E, 0x34, 0x41, 0x7E, 0x32,
		0x59, 0x31, 0x35, 0x30, 0x30, 0x23, 0x3F})

	if !status {
		r.InitStatus = "Failed on sending request"
		return false
	}
	r.InitStatus = "Commiting request"

	status = r.sendCommand(conn, "Commit", []byte{
		0x3F, 0x23, 0x7E, 0x34, 0x42, 0x7E, 0x23, 0x3F})

	r.InitStatus = "OK"
	diag.Info("Sent init command to Growatt inverter.")
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
		diag.Warn(task + " not accepted: " + err2.Error())
		return false
	}
	if size == 0 {
		diag.Warn(task + " not accepted: Empty response.")
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
		diag.Warn(task + " not accepted: Code " + strconv.Itoa(int(first)) + ".")
		return false
	}

	diag.Verbose("Reading size of send command: " + strconv.Itoa(size))
	diag.Verbose(hex.Dump(buffer[0:size]))
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

	diag.Info(fmt.Sprintf("Connecting to %v [%v,8,N,1]", r.device, r.speed))

	r.Status = "Connecting"

	// Open the port.
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	// Make sure to close it later.
	defer conn.Close()

	r.connection = conn
	reading := false

	buffer := make([]byte, 30)
	for {
		r.Status = "Reading since " + r.lastUpdate.Format("15:04:05")
		n, err := conn.Read(buffer)
		if err != nil {
			diag.Warn("Reading failed due to: " + err.Error())
			break
		}

		span := time.Now().Sub(r.lastUpdate)
		r.lastUpdate = time.Now()
		if span > 5*time.Minute {
			diag.Warn("Respawning...")
			r.dataqueue.Clear()
			_ = conn.Close()
			r.connection = nil
			return true
		}

		if !reading {
			reading = true
			diag.Info("Reading started with " + strconv.Itoa(n) + " bytes.")
		}

		// TODO Error because it keeps on reading and getting data. How to stop it?
		// Verbose("Read bytes and pushing: " + strconv.Itoa(n))

		for i := 0; i < n; i++ {
			r.dataqueue.Push(buffer[i])
		}
	}
	return false
}
