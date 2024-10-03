package log

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	FileLogger    *log.Logger
	ConsoleLogger *log.Logger
)

const (
	Reset        = "\033[0m"
	BrightRed    = "\033[1;31m" // Bright Red
	BrightGreen  = "\033[1;32m" // Bright Green
	BrightYellow = "\033[1;33m" // Bright Yellow
	BrightBlue   = "\033[1;34m" // Bright Blue
)

// Init initializes the loggers for file and console
func Init() {
	// Attempt to remove the old log file if it exists
	if err := os.Remove("application.log"); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to delete old log file: %v\n", err)
		return
	}

	file, err := os.OpenFile("application.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}

	// Set up loggers without prefixes
	FileLogger = log.New(file, "", 0)         // No prefix and no flags
	ConsoleLogger = log.New(os.Stdout, "", 0) // No prefix and no flags
}

// currentTimestamp returns the current date and time in the desired format
func currentTimestamp() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

// Info logs informational messages in the specified format
func Info(msg string, args ...interface{}) {
	msgFormatted := fmt.Sprintf("[INFO]   %s - %s 😊", currentTimestamp(), fmt.Sprintf(msg, args...))
	ConsoleLogger.Println(BrightGreen + msgFormatted + Reset)
	FileLogger.Println(msgFormatted)
}

// Warning logs warning messages in the specified format
func Warning(msg string, args ...interface{}) {
	msgFormatted := fmt.Sprintf("[WARNING] %s - %s ⚠️", currentTimestamp(), fmt.Sprintf(msg, args...))
	ConsoleLogger.Println(BrightYellow + msgFormatted + Reset)
	FileLogger.Println(msgFormatted)
}

// Error logs error messages in the specified format
func Error(msg string, args ...interface{}) error {
	msgFormatted := fmt.Sprintf("[ERROR]  %s - %s ❌", currentTimestamp(), fmt.Sprintf(msg, args...))
	ConsoleLogger.Println(BrightRed + msgFormatted + Reset)
	FileLogger.Println(msgFormatted)
	return nil
}

// Fatal logs fatal messages in the specified format and exits the program
func Fatal(msg string, args ...interface{}) {
	msgFormatted := fmt.Sprintf("[FATAL]  %s - %s 💀", currentTimestamp(), fmt.Sprintf(msg, args...))
	ConsoleLogger.Println(BrightRed + msgFormatted + Reset)
	FileLogger.Println(msgFormatted)
	os.Exit(1)
}
