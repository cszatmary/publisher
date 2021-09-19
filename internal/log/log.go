// Package log provides a simple logger.
package log

import (
	"fmt"
	"io"
	"sync"
)

type Logger struct {
	mu           sync.Mutex
	out          io.Writer
	buf          []byte // for accumulating text to write
	debugEnabled bool
}

func New(out io.Writer) *Logger {
	return &Logger{out: out}
}

func (l *Logger) SetDebug(enabled bool) {
	l.debugEnabled = enabled
}

func (l *Logger) log(s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	// Add a trailling newline if there is none
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

func (l *Logger) Printf(format string, v ...interface{}) {
	_ = l.log(fmt.Sprintf(format, v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.debugEnabled {
		_ = l.log(fmt.Sprintf("DEBUG: "+format, v...))
	}
}
