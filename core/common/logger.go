package common

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/google/logger"
)

const (
	NoLogs      int = 0
	DebugLogs   int = 1
	InfoLogs    int = 2
	WarningLogs int = 4
	ErrorLogs   int = 8
)

// ILogger is an interface for loggers
type ILogger interface {
	Debug(messaage string)
	Info(message string)
	Warning(message string)
	Error(message string)
	InitLogger(logLevel int, logFile io.Writer)
}

var singleton *BaseLogger
var once sync.Once

// BaseLog static function returns logger
func BaseLog() *BaseLogger {
	once.Do(func() {
		singleton = newBaseLogger()
	})
	return singleton
}

// BaseLogger is a base implementaion of ILogger
type BaseLogger struct {
	logLevel    int
	initialized bool
}

func newBaseLogger() *BaseLogger {
	b := BaseLogger{15, false}
	return &b
}

// InitLogger used to initialized logger
func (b *BaseLogger) InitLogger(logLevel int) {
	b.logLevel = logLevel
	logger.SetFlags(log.LstdFlags)
	b.initialized = true
}

// Debug used to log debug to console
func (b *BaseLogger) Debug(message string) {
	if b.initialized && b.logLevel&DebugLogs != 0 {
		fmt.Println("[DEBUG] " + message)
	}
}

// Info used to log info to console
func (b *BaseLogger) Info(message string) {
	if b.initialized && b.logLevel&InfoLogs != 0 {
		fmt.Println("[INFO] " + message)
	}
}

// Warning used to log warnings to console
func (b *BaseLogger) Warning(message string) {
	if b.initialized && b.logLevel&WarningLogs != 0 {
		fmt.Println("[WARNING] " + message)
	}
}

// Error used to log errors to console
func (b *BaseLogger) Error(message string) {
	if b.initialized && b.logLevel&ErrorLogs != 0 {
		fmt.Println("[ERROR] " + message)
	}
}
