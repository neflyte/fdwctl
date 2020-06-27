package logger

import (
	"github.com/sirupsen/logrus"
	"strings"
)

var (
	rootLogger *logrus.Logger
)

func init() {
	rootLogger = logrus.StandardLogger()
}

func SetFormat(format string) {
	formatStr := strings.TrimSpace(strings.ToLower(format))
	switch formatStr {
	case "json":
		rootLogger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		rootLogger.SetFormatter(&logrus.TextFormatter{})
	default:
		rootLogger.WithField("function", "SetFormat").Errorf("unknown format: %s", formatStr)
	}
}

func Root() *logrus.Logger {
	return rootLogger
}
