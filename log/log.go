// Package log defines the logging interface used in this library. This
// logging interface is compatible with github.com/apex/log. However, we
// only use the interface in this library. Therefore, it's possible in
// principle for you to use this library with another logger, as long as
// you make such logger implement this interface.
package log

// Logger defines the common interface that a logger should have. It is
// out of the box compatible with `log.Log` in `apex/log`.
type Logger interface {
	// Debug emits a debug message.
	Debug(msg string)

	// Debugf formats and emits a debug message.
	Debugf(format string, v ...interface{})

	// Info emits an informational message.
	Info(msg string)

	// Infof format and emits an informational message.
	Infof(format string, v ...interface{})

	// Warn emits a warning message.
	Warn(msg string)

	// Warnf formats and emits a warning message.
	Warnf(format string, v ...interface{})
}
