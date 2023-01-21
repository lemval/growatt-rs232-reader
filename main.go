// Growatt project main.go
package main

import (
	"flag"
	"fmt"

	"os"
	"strings"

	"growattrr/diag"
	"growattrr/reader"
)

var speed int
var port int
var device string
var action string
var broker string
var topic string
var verbose bool
var delay int
var user string
var credential string

func init() {
	flag.StringVar(&action, "action", "Start", "The action (Start or Init).")
	flag.StringVar(&device, "device", "/dev/ttyUSB0", "The serial port descriptor.")
	flag.StringVar(&broker, "broker", "", "Connect to MQTT broker (e.g. tcp://localhost:1883).")
	flag.StringVar(&topic, "topic", "Growatt", "MQTT topic /solar/<topic>/<item>.")
	flag.StringVar(&user, "user", "", "MQTT user (leave empty to use unauthorized).")
	flag.StringVar(&credential, "password", "", "MQTT password.")
	flag.IntVar(&speed, "baudrate", 9600, "The baud rate of the serial connection.")
	flag.IntVar(&port, "server", 5701, "The server port for the REST service.")
	flag.IntVar(&delay, "delay", 0, "Period (seconds) of delay to publish values on MQTT.")
	flag.BoolVar(&verbose, "v", false, "Activate verbose logging.")
}

var Version = "v1.400"

func main() {
	diag.Info("Starting Growatt Inverter Reader " + Version)

	//	Read the command line arguments
	flag.Parse()

	diag.Verbosive = verbose

	// Initialize the reader
	serialReader := reader.NewReader(device, speed)

	//	Handle the 'init' command to to send the message to
	//	start the logging of the data.

	if strings.Compare("Init", action) == 0 {
		actionInit(serialReader)
	} else if strings.Compare("Start", action) == 0 {
		actionStart(serialReader)
	} else {
		fmt.Printf("\n == ERROR ==============================")
		fmt.Printf("\n    Invalid action '%s'!", action)
		fmt.Printf("\n =======================================")
		fmt.Printf("\n Usage: %s [<options>]", os.Args[0])
		flag.PrintDefaults()
	}
}

func actionInit(reader *reader.Reader) {
	diag.Info("Init requested...")
	status := reader.InitLogger()
	if status {
		diag.Info("Sent. Please restart!")
	} else {
		diag.Warn("Failed. Please retry!")
	}
}

func actionStart(reader *reader.Reader) {

	// Initialize the interpreter and publisher and start all threads to
	// read data, interpret to datagrams and publish as json

	interpreter := NewInterpreter(reader.GetQueue())
	publisher := NewPublisher(delay)

	go reader.StartMonitored()
	go interpreter.start()
	go publisher.start(port)

	publisher.listen(interpreter, reader)
}
