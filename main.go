// Growatt project main.go
package main

import (
	"fmt"
	"time"
)

func main() {
	reader := NewReader()
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
