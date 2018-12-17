// Growatt project main.go
package main

import (
	"flag"
	"fmt"

	"os"
	// "strconv"
	"strings"
	"time"
)

var speed int
var port int
var device string
var action string

func init() {
	flag.StringVar(&action, "action", "Start", "The action (Start or Init).")
	flag.StringVar(&device, "device", "/dev/ttyUSB0", "The serial port descriptor.")
	flag.IntVar(&speed, "baudrate", 9600, "The baud rate of the serial connection.")
	flag.IntVar(&port, "server", 5701, "The server port for the REST service.")
}

func main() {
	//	Read the command line arguments
	flag.Parse()

	// Initialize the reader
	reader := NewReader(device, speed)

	//	Handle the 'init' command to to send the message to
	//	start the logging of the data.

	if strings.Compare("Init", action) == 0 {
		actionInit(reader)
	} else if strings.Compare("Start", action) == 0 {
		actionStart(reader)
	} else {
		fmt.Printf("\n == ERROR ==============================")
		fmt.Printf("\n    Invalid action '%s'!", action)
		fmt.Printf("\n =======================================")
		fmt.Printf("\n Usage: %s [<options>]", os.Args[0])
		flag.PrintDefaults()
	}
}

func actionInit(reader *Reader) {
	Info("Init requested...")
	reader.initLogger(false)
	Info("Sent. Please restart!")
}

func actionStart(reader *Reader) {

	// Initialize the interpreter and publisher and start all threads to
	// read data, interpret to datagrams and publish as json

	interpreter := NewInterpreter(reader.getQueue())
	publisher := new(Publisher)

	go reader.startMonitored()
	go interpreter.start()
	go publisher.start(port)

	sleepInduced := false

	for {
		data := interpreter.pop()
		if data != nil {
			sleepInduced = false
			publisher.updateData(data)
		} else {
			if sleepInduced {
				publisher.updateData(nil)
			}
			// Only sleep if there are no datagrams currently
			time.Sleep(500 * time.Millisecond)
			sleepInduced = true
		}
	}
}

func writeMessage(msg string, ctx string) {
	now := time.Now().Format("15:04")
	fmt.Printf("%s %s %s\n", now, ctx, msg)
}

func Warn(msg string) {
	writeMessage(msg, "[WARN]")
}

func Info(msg string) {
	writeMessage(msg, "[INFO]")
}
