package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(level, format string) {
	log = logrus.New()

	switch format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.SetFormatter(&logrus.TextFormatter{})
	}

	log.SetOutput(os.Stdout)

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		parsedLevel = logrus.InfoLevel
	}
	log.SetLevel(parsedLevel)
}

func Get() *logrus.Logger {
	if log == nil {
		Init("info", "json")
	}
	return log
}

func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Get().WithFields(fields)
}

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}
