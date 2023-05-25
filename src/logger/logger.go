package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func NewLogFile(filename string) (*os.File, error) {
	var file *os.File
	var err error
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		var dir = filepath.Dir(filename)
		os.MkdirAll(dir, os.ModePerm)
		file, err = os.Create(filename)
	} else if err == nil {
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	return file, err
}

type logger struct {
	Loglevel Loglevel
	prefix   string
	File     io.Writer
}

func Newlogger(loglevel Loglevel, w io.Writer, prefix ...string) Logger {
	var l = logger{
		Loglevel: loglevel,
		File:     w,
	}
	if len(prefix) > 0 {
		l.prefix = prefix[0]
	}
	return &l
}

func (l *logger) Critical(err error) {
	l.logLine(CRITICAL, err.Error())
}

// Write an error message, loglevel error
func (l *logger) Error(args ...any) {
	l.logLine(ERROR, fmt.Sprint(args...))
}

// Write an error message, loglevel error
func (l *logger) Errorf(format string, args ...any) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

// Write a warning message, loglevel warning
func (l *logger) Warning(args ...any) {
	l.logLine(WARNING, fmt.Sprint(args...))
}

// Write a warning message, loglevel warning
func (l *logger) Warningf(format string, args ...any) {
	l.log(WARNING, fmt.Sprintf(format, args...))
}

// Write an info message, loglevel info
func (l *logger) Info(args ...any) {
	l.logLine(INFO, fmt.Sprint(args...))
}

// Write an info message, loglevel info
func (l *logger) Infof(format string, args ...any) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

// Write a debug message, loglevel debug
func (l *logger) Debug(args ...any) {
	l.logLine(DEBUG, fmt.Sprint(args...))
}

// Write a debug message, loglevel debug
func (l *logger) Debugf(format string, args ...any) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

// Write a test message, loglevel test
func (l *logger) Test(args ...any) {
	l.logLine(TEST, fmt.Sprint(args...))
}

// Write a test message, loglevel test
func (l *logger) Testf(format string, args ...any) {
	l.log(TEST, fmt.Sprintf(format, args...))
}

func (l *logger) logLine(level Loglevel, msg string) {
	l.log(level, msg+"\n")
}

func (l *logger) log(msgType Loglevel, msg string) {
	if l.Loglevel >= Loglevel(msgType) {
		fmt.Fprintf(l.File, "%s%s", generatePrefix(true, l.prefix, msgType), msg)
	}
}

func generatePrefix(colorized bool, prefix string, level Loglevel) string {
	var msg string
	msg = "[%s%s] "
	msg = fmt.Sprintf(msg, prefix, level.String())
	msg = timestamp(msg)
	if colorized {
		var color = getLogLevelColor(level)
		msg = Colorize(msg, color)
	}
	return msg
}

func timestamp(msg string) string {
	return fmt.Sprintf("%s %s", time.Now().Format("2006-01-02 15:04:05"), msg)
}
