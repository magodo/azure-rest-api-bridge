package log

var logger Logger = &NullLogger{}

func SetLogger(l Logger) {
	logger = l
}

type Logger interface {
	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type NullLogger struct{}

func (n *NullLogger) Trace(msg string, args ...interface{}) {
	return
}

func (n *NullLogger) Debug(msg string, args ...interface{}) {
	return
}

func (n *NullLogger) Info(msg string, args ...interface{}) {
	return
}

func (n *NullLogger) Warn(msg string, args ...interface{}) {
	return
}

func (n *NullLogger) Error(msg string, args ...interface{}) {
	return
}

func Trace(msg string, args ...interface{}) {
	logger.Trace(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	logger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	logger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	logger.Error(msg, args...)
}
