// Growatt project main.go
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	args := os.Args[1:]

	var speed int
	device := "/dev/ttyUSB0"
	if len(args) > 0 && len(args[0]) > 0 {
		device = args[0]
	}
	if len(args) > 1 && len(args[1]) > 0 {
		speed, _ = strconv.Atoi(args[1])
	}

	reader := NewReader(device, speed)

	if len(args) > 2 && strings.Compare("init", args[2]) == 0 {
		Info("Init requested...")
		reader.initLogger(false)
		Info("Sent. Please restart!")
		return
	}

	interpreter := NewInterpreter(reader.getQueue())
	publisher := new(Publisher)

	go reader.startMonitored()
	go interpreter.start()
	go publisher.start()

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
