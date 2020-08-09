package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type elasticHook struct{}

func NewElasticHook() logrus.Hook {
	return new(elasticHook)
}

func (eh *elasticHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (eh *elasticHook) Fire(entry *logrus.Entry) error {
	var levelText string
	// Add log level to the log.level data key
	switch entry.Level {
	case logrus.PanicLevel:
		levelText = "panic"
	case logrus.FatalLevel:
		levelText = "fatal"
	case logrus.ErrorLevel:
		levelText = "error"
	case logrus.WarnLevel:
		levelText = "warn"
	case logrus.InfoLevel:
		levelText = "info"
	case logrus.DebugLevel:
		levelText = "debug"
	case logrus.TraceLevel:
		levelText = "trace"
	default:
		return fmt.Errorf("unknown log level %s", entry.Level)
	}
	if entry.Data["log"] == nil {
		entry.Data["log"] = make(map[string]interface{})
	}
	logData := entry.Data["log"].(map[string]interface{})
	logData["level"] = levelText
	entry.Data["log"] = logData
	// All done
	return nil
}
