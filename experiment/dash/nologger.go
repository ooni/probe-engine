package dash

type noLogger struct{}

func (noLogger) Debug(msg string) {}

func (noLogger) Debugf(format string, v ...interface{}) {}
