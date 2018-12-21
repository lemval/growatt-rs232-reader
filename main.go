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
	Info("Starting Growatt Inverter Reader v1.001")

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
	status := reader.initLogger()
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
	publisher := NewPublisher()

	go reader.startMonitored()
	go interpreter.start()
	go publisher.start(port)

	publisher.listen(interpreter, reader)
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
		writeMessage(msg, "[----]")
	}
}
