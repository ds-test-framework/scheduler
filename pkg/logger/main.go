package logger

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var logFile *os.File = nil

// Debug logs a debug message
func Debug(s string) {
	log.Println(s)
}

// Fatal logs the message and exits with non-zero exit code
func Fatal(s string) {
	log.Fatalln(s)
}

// Init initializes the default logger with a log path if specified
func Init(c *viper.Viper) {
	log.SetFlags(log.Ldate | log.Ltime)
	path := c.GetString("path")
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	logFile = file
	log.SetOutput(file)
}

// Destroy closes the log file
func Destroy() {
	if logFile != nil {
		logFile.Close()
	}
}
