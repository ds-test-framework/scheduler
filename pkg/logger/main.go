package logger

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

var logFile *os.File = nil

func Debug(s string) {
	log.Println(s)
}

func Fatal(s string) {
	log.Fatalln(s)
}

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

func Destroy() {
	if logFile != nil {
		logFile.Close()
	}
}
