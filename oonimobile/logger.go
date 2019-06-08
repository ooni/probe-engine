package oonimobile

import "fmt"

// LogMessage is a log message
type LogMessage struct {
	// Level indicates the log level
	Level string

	// Message is the log message
	Message string
}

type channelLogger struct {
	out chan<- *LogMessage
}

func (cl *channelLogger) Debug(msg string) {
	cl.out <- &LogMessage{"DEBUG", msg}
}

func (cl *channelLogger) Debugf(format string, v ...interface{}) {
	cl.out <- &LogMessage{"DEBUG", fmt.Sprintf(format, v...)}
}

func (cl *channelLogger) Info(msg string) {
	cl.out <- &LogMessage{"INFO", msg}
}

func (cl *channelLogger) Infof(format string, v ...interface{}) {
	cl.out <- &LogMessage{"INFO", fmt.Sprintf(format, v...)}
}

func (cl *channelLogger) Warn(msg string) {
	cl.out <- &LogMessage{"WARNING", msg}
}

func (cl *channelLogger) Warnf(format string, v ...interface{}) {
	cl.out <- &LogMessage{"WARNING", fmt.Sprintf(format, v...)}
}
