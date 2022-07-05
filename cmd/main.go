package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var gitSHA string //nolint

func main() {
	var one int
	var two int64
	var three string
	var four []string
	var five bool

	// logging
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	log.SetOutput(os.Stdout)
	logger := logrus.NewEntry(log.WithFields(logrus.Fields{
		"app":   "awesomeProject",
		"env":   os.Getenv("ENV"),
		"build": gitSHA,
	}))

	// setup api

}
