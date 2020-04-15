package dash

// NoLogger is a fake logger.
type noLogger struct{}

// Debug emits a debug message.
func (noLogger) Debug(msg string) {}

// Debugf formats and emits a debug message.
func (noLogger) Debugf(format string, v ...interface{}) {}
