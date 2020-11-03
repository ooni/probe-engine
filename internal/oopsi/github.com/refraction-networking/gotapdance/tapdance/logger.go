package tapdance

import (
	"fmt"
	"github.com/ooni/probe-engine/internal/oopsi/github.com/sirupsen/logrus"
	"sync"
)

// implements interface logrus.Formatter
type formatter struct {
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("[%s] %s\n", entry.Time.Format("15:04:05"), entry.Message)), nil
}

var logrusLogger *logrus.Logger
var initLoggerOnce sync.Once

// Logger is an access point for TapDance-wide logger
func Logger() *logrus.Logger {
	initLoggerOnce.Do(func() {
		logrusLogger = logrus.New()
		logrusLogger.Formatter = new(formatter)
		logrusLogger.Level = logrus.InfoLevel
		// logrusLogger.Level = logrus.DebugLevel

		// buildInfo const will be overwritten by CI with `sed` for test builds
		// if not overwritten -- this is a NO-OP
		const buildInfo = ""
		if len(buildInfo) > 0 {
			logrusLogger.Infof("Running gotapdance build %s", buildInfo)
		}
	})
	return logrusLogger
}
