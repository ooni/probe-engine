package oonimkall

import (
	"fmt"
)

// chanLogger is a logger targeting a channel
type chanLogger struct {
	emitter    *eventEmitter
	hasdebug   bool
	hasinfo    bool
	haswarning bool
	out        chan<- *eventRecord
	settings   *settingsRecord
}

// Debug implements Logger.Debug
func (cl *chanLogger) Debug(msg string) {
	if cl.hasdebug {
		cl.emitter.Emit("log", eventValue{
			LogLevel: "DEBUG",
			Message:  msg,
		})
	}
}

// Debugf implements Logger.Debugf
func (cl *chanLogger) Debugf(format string, v ...interface{}) {
	if cl.hasdebug {
		cl.Debug(fmt.Sprintf(format, v...))
	}
}

// Info implements Logger.Info
func (cl *chanLogger) Info(msg string) {
	if cl.hasinfo {
		cl.emitter.Emit("log", eventValue{
			LogLevel: "INFO",
			Message:  msg,
		})
	}
}

// Infof implements Logger.Infof
func (cl *chanLogger) Infof(format string, v ...interface{}) {
	if cl.hasinfo {
		cl.Info(fmt.Sprintf(format, v...))
	}
}

// Warn implements Logger.Warn
func (cl *chanLogger) Warn(msg string) {
	if cl.haswarning {
		cl.emitter.Emit("log", eventValue{
			LogLevel: "WARNING",
			Message:  msg,
		})
	}
}

// Warnf implements Logger.Warnf
func (cl *chanLogger) Warnf(format string, v ...interface{}) {
	if cl.haswarning {
		cl.Warn(fmt.Sprintf(format, v...))
	}
}

// newChanLogger creates a new ChanLogger instance.
func newChanLogger(
	emitter *eventEmitter, settings *settingsRecord,
	out chan<- *eventRecord,
) *chanLogger {
	cl := &chanLogger{
		emitter:  emitter,
		out:      out,
		settings: settings,
	}
	switch settings.LogLevel {
	case "DEBUG", "DEBUG2":
		cl.hasdebug = true
		fallthrough
	case "INFO":
		cl.hasinfo = true
		fallthrough
	case "ERR", "WARNING":
		fallthrough
	default:
		cl.haswarning = true
	}
	return cl
}
