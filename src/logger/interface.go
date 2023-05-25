package logger

type Logger interface {
	Critical(err error)
	Error(args ...any)
	Errorf(format string, args ...any)
	Warning(args ...any)
	Warningf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Debug(args ...any)
	Debugf(format string, args ...any)
	Test(args ...any)
	Testf(format string, args ...any)
}
