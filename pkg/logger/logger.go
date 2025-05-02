package logger

import (
	// "fmt"
	"log"
	"os"
	// "time"
)

// Logger represents a simple logger
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs info messages
func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

// Infof logs formatted info messages
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs error messages
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs debug messages
func (l *Logger) Debug(v ...interface{}) {
	l.debugLogger.Println(v...)
}

// Debugf logs formatted debug messages
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}
