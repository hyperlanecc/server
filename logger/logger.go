package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init(logFile string, level string) {
	Log = logrus.New()

	// 自动创建父目录
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		Log.Out = os.Stdout
	} else {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			Log.Out = os.Stdout
		} else {
			Log.Out = file
		}
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Log.SetLevel(logLevel)
}
