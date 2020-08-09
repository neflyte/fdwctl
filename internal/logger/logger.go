/*
Package logger handles application logging
*/
package logger

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"strings"
)

const (
	// TraceLevel represents the TRACE logging level
	TraceLevel = "trace"
	// DebugLevel represents the DEBUG logging level
	DebugLevel = "debug"
	// InfoLevel represents the INFO logging level
	InfoLevel = "info"
	// WarnLevel represents the WARN logging level
	WarnLevel = "warn"
	// ErrorLevel represents the ERROR logging level
	ErrorLevel = "error"
	// FatalLevel represents the FATAL logging level
	FatalLevel = "fatal"
	// PanicLevel represents the PANIC logging level
	PanicLevel = "panic"

	// JSONFormat represents the JSON logging format
	JSONFormat = "json"
	// TextFormat represents the text logging format
	TextFormat = "text"
	// ECSFormat represents the Elasticstack Common Schema (ECS) JSON logging format
	ECSFormat = "elastic"
)

var (
	// rootLogger is the singleton application logger
	rootLogger *logrus.Logger
	// rootLoggerFields is the map of fields to include in every log message
	rootLoggerFields = make(logrus.Fields)
)

// SetFormat configures the logger message format
func SetFormat(format string) {
	formatStr := strings.TrimSpace(strings.ToLower(format))
	switch formatStr {
	case ECSFormat:
		Root().SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "@timestamp",
				logrus.FieldKeyMsg:  "message",
			},
		})
		Root().AddHook(NewElasticHook())
	case JSONFormat:
		Root().SetFormatter(&logrus.JSONFormatter{})
	case TextFormat:
		Root().SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		Root().WithField("function", "SetFormat").Errorf("unknown format: %s", formatStr)
	}
}

// SetLevel configures the logger logging level
func SetLevel(level string) {
	levelStr := strings.TrimSpace(strings.ToLower(level))
	switch levelStr {
	case TraceLevel:
		Root().SetLevel(logrus.TraceLevel)
	case DebugLevel:
		Root().SetLevel(logrus.DebugLevel)
	case InfoLevel:
		Root().SetLevel(logrus.InfoLevel)
	case WarnLevel:
		Root().SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		Root().SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		Root().SetLevel(logrus.FatalLevel)
	case PanicLevel:
		Root().SetLevel(logrus.PanicLevel)
	default:
		Root().WithField("function", "SetLevel").Errorf("unknown level: %s", level)
	}
}

// Root returns the singleton application logger
func Root() *logrus.Logger {
	if rootLogger == nil {
		rootLogger = logrus.StandardLogger()
		rootLogger.SetLevel(logrus.DebugLevel)
		// Get machine hostname
		hostname, err := os.Hostname()
		if err != nil {
			Root().WithField("function", "SetFormat").
				Errorf("error getting hostname: %s", err)
			hostname = ""
		}
		// If the hostname is non-empty, add it as a logger field
		if hostname != "" {
			rootLoggerFields["host"] = map[string]interface{}{
				"name": hostname,
			}
		}
	}
	return rootLogger
}

// Log returns a logrus FieldLogger including the root fields and an optional context
func Log(args ...interface{}) logrus.FieldLogger {
	var entry *logrus.Entry
	for _, arg := range args {
		switch obj := arg.(type) {
		case context.Context:
			entry = rootLogger.WithContext(obj)
		}
	}
	if entry == nil {
		entry = rootLogger.WithFields(rootLoggerFields)
	} else {
		entry = entry.WithFields(rootLoggerFields)
	}
	return entry
}

// ErrorfAsError logs an Error message to the supplied logger and then returns a
// new error object initialized with the message. The message is formatted with
// fmt.Sprintf() before passing to the logger and the error object.
func ErrorfAsError(log logrus.FieldLogger, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	log.Error(message)
	return errors.New(message)
}

// SanitizedURLString returns a parsed URL string with user credentials removed
func SanitizedURLString(urlWithCreds string) string {
	log := Log().
		WithField("function", "SanitizedURLString")
	clone, err := url.Parse(urlWithCreds)
	if err != nil {
		log.Errorf("unable to clone url: %s", err)
		return urlWithCreds
	}
	if clone.User != nil {
		clone.User = url.User(clone.User.Username())
	}
	return clone.String()
}
