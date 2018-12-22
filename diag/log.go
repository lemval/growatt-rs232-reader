package diag

import (
	"fmt"
	"time"
)

var Verbosive = false

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
	if Verbosive {
		writeMessage(msg, "[----]")
	}
}
