package core

import (
	"os"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Logger *logrus.Logger

func init() {
	logger := logrus.New()
	logger = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.InfoLevel,
		Formatter: &prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
		},
	}
	Logger = logger
}
