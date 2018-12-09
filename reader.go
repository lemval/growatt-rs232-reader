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
}

func NewReader(device string, speed int) *Reader {
	r := new(Reader)
	r.device = device
	r.speed = uint(speed)
	if r.speed == 0 {
		r.speed = 9600
	}
	r.dataqueue = NewQueue()

	return r
}

func (r *Reader) getQueue() *Queue {
	return r.dataqueue
}

func (r *Reader) sendInitCommand(conn io.ReadWriteCloser) {
	initString1 := []byte{0x3F, 0x23, 0x7E, 0x34, 0x41, 0x7E, 0x32, 0x59, 0x35, 0x30, 0x30, 0x30, 0x23, 0x3F}
	initString2 := []byte{0x3F, 0x23, 0x7E, 0x34, 0x42, 0x7E, 0x23, 0x3F}

	log.Println("[WARN] serial.sendInitCommand sending")

	_, err1 := conn.Write(initString1)
	if err1 != nil {
		log.Fatalf("[ERROR] serial.sendInitCommand: %v", err1)
	}

	time.Sleep(100 * time.Millisecond)

	_, err2 := conn.Write(initString2)
	if err2 != nil {
		log.Fatalf("[ERROR] serial.sendInitCommand: %v", err2)
	}
}

func (r *Reader) initLogger() {
	options := serial.OpenOptions{
		PortName:          r.device,
		BaudRate:          r.speed,
		DataBits:          8,
		StopBits:          1,
		ParityMode:        0,
		MinimumReadSize:   30,
		RTSCTSFlowControl: false,
	}
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	defer conn.Close()
	r.sendInitCommand(conn)
}

func (r *Reader) start() {
	options := serial.OpenOptions{
		PortName:          r.device,
		BaudRate:          r.speed,
		DataBits:          8,
		StopBits:          1,
		ParityMode:        0,
		MinimumReadSize:   30,
		RTSCTSFlowControl: false,
	}

	fmt.Printf("Connecting to %v [%v,8,N,1]\n", r.device, r.speed)

	// Open the port.
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("[ERROR] serial.Open: %v", err)
	}
	r.connection = conn

	// Make sure to close it later.
	defer r.connection.Close()

	var buffer []byte
	buffer = make([]byte, 30)
	zeroCounter := 0
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatalf("[ERROR] serial.Read: %v", err)
		}
		for i := 0; i < n; i++ {
			if buffer[i] == 0x00 {
				zeroCounter = zeroCounter + 1
				if zeroCounter > 200 {
					r.sendInitCommand(conn)
				}
			} else {
				zeroCounter = 0
			}
			r.dataqueue.Push(buffer[i])
		}
	}
}
