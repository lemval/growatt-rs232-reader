// Growatt project main.go
package main

import (
	"os"
	"strconv"
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
	interpreter := NewInterpreter(reader.getQueue())
	publisher := new(Publisher)

	go reader.start()
	go interpreter.start()
	go publisher.start()

	for {
		data := interpreter.pop()
		if data != nil {
			publisher.updateData(data)
		} else {
			// Only sleep if there are no datagrams currently
			time.Sleep(500 * time.Millisecond)
		}
	}

}
