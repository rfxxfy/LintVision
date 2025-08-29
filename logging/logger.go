package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type outputFormat int

const (
	TextFormat outputFormat = iota
	JSONFormat
)

type Logger struct {
	mu         sync.Mutex
	out        io.Writer
	level      Level
	timeFormat string
	format     outputFormat
	caller     bool
}

var std = New(os.Stdout, INFO)

func New(out io.Writer, lvl Level) *Logger {
	return &Logger{
		out:        out,
		level:      lvl,
		timeFormat: time.RFC3339,
		format:     TextFormat,
		caller:     false,
	}
}

func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

func SetLevel(lvl Level) {
	std.SetLevel(lvl)
}

func SetFormat(fmtOut outputFormat) {
	std.SetFormat(fmtOut)
}

func SetCaller(on bool) {
	std.SetCaller(on)
}

func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	l.out = out
	l.mu.Unlock()
}

func (l *Logger) SetLevel(lvl Level) {
	l.mu.Lock()
	l.level = lvl
	l.mu.Unlock()
}

func (l *Logger) SetFormat(fmtOut outputFormat) {
	l.mu.Lock()
	l.format = fmtOut
	l.mu.Unlock()
}

func (l *Logger) SetCaller(on bool) {
	l.mu.Lock()
	l.caller = on
	l.mu.Unlock()
}

func (l *Logger) logMsg(lvl Level, msg string) {
	if lvl < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().Format(l.timeFormat)
	if l.format == JSONFormat {
		record := map[string]interface{}{
			"time":  ts,
			"level": lvl.String(),
			"msg":   msg,
		}
		if l.caller {
			if pc, file, line, ok := runtime.Caller(3); ok {
				fn := runtime.FuncForPC(pc)
				record["caller"] = fmt.Sprintf("%s:%d %s", file, line, fn.Name())
			}
		}
		if b, err := json.Marshal(record); err == nil {
			fmt.Fprintln(l.out, string(b))
		} else {
			fmt.Fprintf(l.out, `{"time":"%s","level":"%s","msg":"%s","error":"%v"}`+"\n",
				ts, lvl.String(), msg, err)
		}
	} else {
		header := fmt.Sprintf("[%s] [%s]", ts, lvl.String())
		entry := header + " " + msg
		if l.caller {
			if pc, file, line, ok := runtime.Caller(3); ok {
				fn := runtime.FuncForPC(pc)
				entry += fmt.Sprintf(" (%s:%d %s)", file, line, fn.Name())
			}
		}
		fmt.Fprintln(l.out, entry)
	}

	if lvl == FATAL {
		os.Exit(1)
	}
}

func Debug(format string, args ...interface{}) {
	std.logMsg(DEBUG, fmt.Sprintf(format, args...))
}

func Info(format string, args ...interface{}) {
	std.logMsg(INFO, fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	std.logMsg(WARN, fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	std.logMsg(ERROR, fmt.Sprintf(format, args...))
}

func Fatal(format string, args ...interface{}) {
	std.logMsg(FATAL, fmt.Sprintf(format, args...))
}
