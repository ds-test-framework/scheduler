package log

import "github.com/sirupsen/logrus"

func DummyLogger() *Logger {
	logger := logrus.New()
	return &Logger{
		entry: logrus.NewEntry(logger),
		file:  nil,
	}
}
