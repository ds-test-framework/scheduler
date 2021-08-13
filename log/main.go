package log

import (
	"os"

	"github.com/ds-test-framework/scheduler/config"
	"github.com/sirupsen/logrus"
)

var DefaultLogger *Logger

type LogParams map[string]interface{}

type Logger struct {
	entry *logrus.Entry

	file *os.File
}

func NewLogger(c config.LogConfig) *Logger {
	l := logrus.New()
	if c.Format == "json" {
		l.SetFormatter(&logrus.JSONFormatter{})
	}

	var file *os.File

	if c.Path != "" {
		file, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			l.SetOutput(file)
		}
	}
	return &Logger{
		entry: logrus.NewEntry(l),
		file:  file,
	}
}

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

func With(params LogParams) *Logger {
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

func (l *Logger) With(params LogParams) *Logger {
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
func Init(c config.LogConfig) {
	DefaultLogger = NewLogger(c)
	DefaultLogger.SetLevel(c.Level)
}

// Destroy closes the log file
func Destroy() {
	DefaultLogger.Destroy()
}
