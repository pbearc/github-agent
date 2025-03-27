package common

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is a wrapper around logrus.Logger
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new Logger instance
func NewLogger() *Logger {
	log := logrus.New()
	
	// Set formatter
	log.SetFormatter(&logrus.JSONFormatter{})
	
	// Set output
	log.SetOutput(os.Stdout)
	
	// Set level based on environment
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.DebugLevel)
	}
	
	return &Logger{log}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}