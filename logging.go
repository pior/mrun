package mrun

import (
	"log"
)

// Logger is the contract our logger
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type standardLogger struct{}

func (*standardLogger) Infof(format string, args ...interface{}) {
	log.Printf("INFO "+format, args...)
}

func (*standardLogger) Warnf(format string, args ...interface{}) {
	log.Printf("WARN "+format, args...)
}
