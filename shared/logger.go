package shared

import (
	"github.com/sirupsen/logrus"
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
	log.Formatter = &logrus.JSONFormatter{}
	log.SetLevel(logrus.InfoLevel)

	return &Logger{
		logger: log,
	}
}

func (l *Logger) Info(message string, fields ...string) {
	l.getEntry(fields).Info(message)
}

func (l *Logger) Warn(message string, fields ...string) {
	l.getEntry(fields).Warn(message)
}

func (l *Logger) Fatal(message string, fields ...string) {
	l.getEntry(fields).Fatal(message)
}

func (l *Logger) getEntry(fields []string) *logrus.Entry {
	e := logrus.NewEntry(l.logger)
	for i := 0; i < len(fields)-1; i += 2 {
		e = e.WithField(fields[i], fields[i+1])
	}
	return e
}
