package logging

import (
	"fmt"
	"os"
	"time"
)

var logFile *os.File

// Init log file
func InitLog(path string) {
	var err error
	logFile, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[-] Failed to open log file:", err)
		os.Exit(1)
	}
}

// Write a log line with timestamp
func LogLine(msg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)
	fmt.Print(line)
	if logFile != nil {
		logFile.WriteString(line)
	}
}

// Close log file on exit
func CloseLog() {
	if logFile != nil {
		logFile.Close()
	}
}
