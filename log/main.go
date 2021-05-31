package log

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var DefaultLogger *Logger

type Logger struct {
	entry *logrus.Entry

	file *os.File
}

func NewLogger(c *viper.Viper) *Logger {
	l := logrus.New()
	if c.GetString("format") == "json" {
		l.SetFormatter(&logrus.JSONFormatter{})
	}
	path := c.GetString("path")

	var file *os.File

	if path != "" {
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			l.SetOutput(file)
		}
	}
	return &Logger{
		entry: logrus.NewEntry(l),
		file:  file,
	}
}

var logFile *os.File = nil

// Debug logs a debug message
func Debug(s string) {
	DefaultLogger.Debug(s)
}

// Fatal logs the message and exits with non-zero exit code
func Fatal(s string) {
	DefaultLogger.Fatal(s)
}

func Info(s string) {
	DefaultLogger.Info(s)
}

func Warn(s string) {
	DefaultLogger.Warn(s)
}

func Error(s string) {
	DefaultLogger.Error(s)
}

func With(params map[string]string) *Logger {
	return DefaultLogger.With(params)
}

func SetLevel(l string) {
	DefaultLogger.SetLevel(l)
}

// Debug logs a debug message
func (l *Logger) Debug(s string) {
	l.entry.Debug(s)
}

// Fatal logs the message and exits with non-zero exit code
func (l *Logger) Fatal(s string) {
	l.entry.Fatal(s)
}

func (l *Logger) Info(s string) {
	l.entry.Info(s)
}

func (l *Logger) Warn(s string) {
	l.entry.Warn(s)
}

func (l *Logger) Error(s string) {
	l.entry.Error(s)
}

func (l *Logger) With(params map[string]string) *Logger {
	fields := logrus.Fields{}
	for k, v := range params {
		fields[k] = v
	}

	entry := l.entry.WithFields(fields)
	return &Logger{
		entry: entry,
		file:  nil,
	}
}

func (l *Logger) SetLevel(level string) {
	levelL, err := logrus.ParseLevel(level)
	if err != nil {
		return
	}
	l.entry.Logger.SetLevel(levelL)
}

func (l *Logger) Destroy() {
	if l.file != nil {
		l.file.Close()
	}
}

// Init initializes the default logger with a log path if specified
func Init(c *viper.Viper, level string) {
	DefaultLogger = NewLogger(c)
	DefaultLogger.SetLevel(level)
}

// Destroy closes the log file
func Destroy() {
	DefaultLogger.Destroy()
}
