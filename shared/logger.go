package shared

import (
	"github.com/sirupsen/logrus"
	"os"
	"fmt"
)

type LoggerInterface interface {
	Info(message string, fields ...string)
	Warn(message string, fields ...string)
	Fatal(message string, fields ...string)
}

type Logger struct {
	logger *logrus.Logger
}

func NewLogger() *Logger {
	log := logrus.New()
	log.Out = os.Stdout
	log.Formatter = &logrus.JSONFormatter{}
	log.Level = logrus.InfoLevel

	return &Logger{
		logger: log,
	}
}

func (l *Logger) Info(message string, fields ...string) {
	//l.getEntry(fields).Info(message)
	fmt.Println(message)
}

func (l *Logger) Warn(message string, fields ...string) {
	//l.getEntry(fields).Info(message)
	fmt.Println(message)
}

func (l *Logger) Fatal(message string, fields ...string) {
	//l.getEntry(fields).Info(message)
	fmt.Println(message)
}

func (l *Logger) getEntry(fields []string) *logrus.Entry {
	e := &logrus.Entry{}
	for i:=0; i<len(fields)-1; i++ {
		e = l.logger.WithFields(logrus.Fields{
			fields[i]: fields[i+1],
		})
	}
	return e.WithField("foo", "bar")
}