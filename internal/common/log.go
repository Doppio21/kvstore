package common

import "github.com/sirupsen/logrus"

func NewLogger() *logrus.Logger {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
	})

	return logrus.StandardLogger()
}
