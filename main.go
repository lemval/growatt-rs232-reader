// Growatt project main.go
package main

import (
	"fmt"
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

	go reader.start()
	go interpreter.start()

	for {
		data := interpreter.pop()
		if data != nil {
			fmt.Println("Got: " + data.String())
		}
		time.Sleep(100 * time.Millisecond)
	}

}
