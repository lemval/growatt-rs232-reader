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
var verbose bool

func init() {
	flag.StringVar(&action, "action", "Start", "The action (Start or Init).")
	flag.StringVar(&device, "device", "/dev/ttyUSB0", "The serial port descriptor.")
	flag.IntVar(&speed, "baudrate", 9600, "The baud rate of the serial connection.")
	flag.IntVar(&port, "server", 5701, "The server port for the REST service.")
	flag.BoolVar(&verbose, "v", false, "Activate verbose logging.")
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
	status := reader.initLogger(false)
	if status {
		Info("Sent. Please restart!")
	} else {
		Warn("Failed. Please retry!")
	}
}

func actionStart(reader *Reader) {

	// Initialize the interpreter and publisher and start all threads to
	// read data, interpret to datagrams and publish as json

	interpreter := NewInterpreter(reader.getQueue())
	publisher := new(Publisher)

	go reader.startMonitored()
	go interpreter.start()
	go publisher.start(port)

	for {
		data := interpreter.pop()
		if data != nil {
			Verbose("Valid datagram: " + data.Status)
			publisher.updateData(data)
		} else {
			publisher.updateData(nil)
			// Only sleep if there are no datagrams currently
			time.Sleep(10 * time.Second)
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

func Verbose(msg string) {
	if verbose {
		writeMessage(msg, "[DEBUG]")
	}
}
