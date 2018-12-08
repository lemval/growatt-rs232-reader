// reader
package main

import (
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

type Reader struct {
	connection io.ReadWriteCloser
	dataqueue  *Queue
}

func NewReader() *Reader {
	r := new(Reader)
	r.dataqueue = NewQueue()

	return r
}

func (r *Reader) getQueue() *Queue {
	return r.dataqueue
}

func (r *Reader) start() {
	options := serial.OpenOptions{
		PortName: "/dev/ttyUSB0",
		BaudRate: 9600,
		DataBits: 8,
		StopBits: 1,
	}

	// Open the port.
	conn, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	r.connection = conn

	// Make sure to close it later.
	defer r.connection.Close()

	var buffer []byte
	buffer = make([]byte, 30)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatalf("serial.Read: %v", err)
		}
		log.Printf("serial.Read: %d bytes", n)
		for i := 0; i < n; i++ {
			r.dataqueue.Push(buffer[i])
		}
	}
}
