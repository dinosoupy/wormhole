/*
Copyright © 2022 Anish Basu

*/
package main

import (
	"os"

	"github.com/dinosoupy/wormhole/cmd"
	log "github.com/sirupsen/logrus"
)

func setupLogger() {
	log.SetOutput(os.Stdout)

	logLevel := log.WarnLevel

	if lvl, ok := os.LookupEnv("GFILE_LOG"); ok {
		switch lvl {
		case "TRACE":
			logLevel = log.TraceLevel
		case "DEBUG":
			logLevel = log.DebugLevel
		case "INFO":
			logLevel = log.InfoLevel
		case "WARN":
			logLevel = log.WarnLevel
		case "PANIC":
			logLevel = log.PanicLevel
		case "ERROR":
			logLevel = log.ErrorLevel
		case "FATAL":
			logLevel = log.FatalLevel
		}
	}
	log.SetLevel(logLevel)
}

func init() {
	setupLogger()
}


func main() {
	cmd.Execute()
}
