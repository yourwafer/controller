package config

import (
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func init() {
	var logFilePath string
	if ApplicationConfig.ServerType == "server" {
		logFilePath = "./logs/server.log"
	} else {
		logFilePath = "./logs/agent.log"
	}
	const rotationHour = 24
	const maxAgeHour = 72
	const timeStampFormat = "2016-01-02 15:04:05.000"
	switch ApplicationConfig.LogLevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: timeStampFormat,
	})
	logrus.SetOutput(ioutil.Discard)
	fileWriter, err := rotatelogs.New(
		logFilePath+".%Y-%m-%d",
		rotatelogs.WithLinkName(logFilePath),
		rotatelogs.WithRotationTime(time.Hour*rotationHour),
		rotatelogs.WithMaxAge(time.Hour*maxAgeHour),
	)
	if err != nil {
		panic(err)
	}
	var accessWriter io.Writer = fileWriter
	if ApplicationConfig.ConsoleOutput {
		accessWriter = io.MultiWriter(accessWriter, os.Stderr)
	}
	logrus.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.TraceLevel: accessWriter,
			logrus.DebugLevel: accessWriter,
			logrus.InfoLevel:  accessWriter,
			logrus.WarnLevel:  accessWriter,
			logrus.ErrorLevel: accessWriter,
			logrus.FatalLevel: accessWriter,
			logrus.PanicLevel: accessWriter,
		},
		&logrus.TextFormatter{
			TimestampFormat: timeStampFormat,
		},
	))
	//
	//logrus.WithFields(logrus.Fields{
	//	"type": ApplicationConfig.ServerType,
	//})
	//logrus.SetReportCaller(true)
}

// Discard is an Writer on which all Write calls succeed
// without doing anything.
var LogrusWriter io.Writer = discard{}

type discard struct{}

func (discard) Write(p []byte) (int, error) {
	logrus.Info(string(p))
	return len(p), nil
}
