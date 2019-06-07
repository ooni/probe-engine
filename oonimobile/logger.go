package oonimobile

import (
	"fmt"
	"strings"
)

type stringLogger struct {
	Builder strings.Builder
}

func (sl *stringLogger) SaveLogsInto(s *string) {
	*s = sl.Builder.String()
}

func (sl *stringLogger) log(msg string) {
	sl.Builder.WriteString(msg)
	sl.Builder.WriteString("\n")
}

func (sl *stringLogger) logf(format string, v ...interface{}) {
	sl.log(fmt.Sprintf(format, v...))
}

func (sl *stringLogger) Debug(msg string) {
	sl.log(msg)
}

func (sl *stringLogger) Debugf(format string, v ...interface{}) {
	sl.logf(format, v...)
}

func (sl *stringLogger) Info(msg string) {
	sl.log(msg)
}

func (sl *stringLogger) Infof(format string, v ...interface{}) {
	sl.logf(format, v...)
}

func (sl *stringLogger) Warn(msg string) {
	sl.log(msg)
}

func (sl *stringLogger) Warnf(format string, v ...interface{}) {
	sl.logf(format, v...)
}
