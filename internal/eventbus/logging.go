package eventbus

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
}

func logError(system, topic string, err error) {
	log.WithFields(logrus.Fields{
		"system": system,
		"topic":  topic,
		"error":  err,
	}).Error("Error occurred")
}

func logSuccess(system, topic string, message interface{}) {
	log.WithFields(logrus.Fields{
		"system":  system,
		"topic":   topic,
		"message": message,
	}).Info("Message produced successfully")
}

func logWarning(system, topic string, message interface{}) {
	log.WithFields(logrus.Fields{
		"system":  system,
		"topic":   topic,
		"message": message,
	}).Warn("Warning")
}
