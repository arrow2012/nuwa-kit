package log

import (
	"strings"
)

// Writer is an io.Writer that writes to the zap logger
type Writer struct {
	LogFunc func(msg string, fields ...interface{})
}

func (w *Writer) Write(p []byte) (n int, err error) {
	msg := string(p)
	// Remove trailing newline
	msg = strings.TrimSuffix(msg, "\n")

	// Default to Info or use configured Level via global Log?
	// But here we simply redirect to our Info or Debug.
	// Gin debug logs are usually Info level in terms of "Output".
	Info(msg)
	return len(p), nil
}

// ErrorWriter writes to Error level
type ErrorWriter struct{}

func (w *ErrorWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	msg = strings.TrimSuffix(msg, "\n")
	Error(msg)
	return len(p), nil
}

// NewGinWriter returns an io.Writer that logs via zap Info
func NewGinWriter() *Writer {
	return &Writer{}
}

// NewGinErrorWriter returns an io.Writer that logs via zap Error
func NewGinErrorWriter() *ErrorWriter {
	return &ErrorWriter{}
}
