package logger

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	TraceLevel = "trace"
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
	PanicLevel = "panic"
	JSONFormat = "json"
	TextFormat = "text"
)

var (
	rootLogger *logrus.Logger
)

func init() {
	rootLogger = logrus.StandardLogger()
	rootLogger.SetLevel(logrus.DebugLevel)
}

func SetFormat(format string) {
	formatStr := strings.TrimSpace(strings.ToLower(format))
	switch formatStr {
	case JSONFormat:
		rootLogger.SetFormatter(&logrus.JSONFormatter{})
	case TextFormat:
		rootLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		rootLogger.WithField("function", "SetFormat").Errorf("unknown format: %s", formatStr)
	}
}

func SetLevel(level string) {
	levelStr := strings.TrimSpace(strings.ToLower(level))
	switch levelStr {
	case TraceLevel:
		rootLogger.SetLevel(logrus.TraceLevel)
	case DebugLevel:
		rootLogger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		rootLogger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		rootLogger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		rootLogger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		rootLogger.SetLevel(logrus.FatalLevel)
	case PanicLevel:
		rootLogger.SetLevel(logrus.PanicLevel)
	default:
		rootLogger.WithField("function", "SetLevel").Errorf("unknown level: %s", level)
	}
}

func Root() *logrus.Logger {
	return rootLogger
}

// ErrorfAsError logs an Error message to the supplied logger and then returns a
// new error object initialized with the message. The message is formatted with
// fmt.Sprintf() before passing to the logger and the error object.
func ErrorfAsError(log logrus.FieldLogger, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	log.Error(message)
	return errors.New(message)
}
