package util

import "github.com/go-kit/kit/log"

// LogWriter implements io.Writer and forwards messages to an underlying
// log.Logger.
type LogWriter struct {
	Log log.Logger
}

// Write writes p as a string to the underlying logger.
func (w *LogWriter) Write(p []byte) (n int, err error) {
	err = w.Log.Log("msg", string(p))
	return len(p), err
}
