package main

import (
	"os"

	"github.com/op/go-logging"
)

/// Loggign configuration
var log = logging.MustGetLogger("hiver")
var logFormat = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{callpath} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func loggingInit() {
	// Create backend for logs output.
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	// Set log format.
	backendFormatter := logging.NewBackendFormatter(backend, logFormat)
	// Set the backends to be used and set logging level.
	backendFormatterLeveled := logging.AddModuleLevel(backendFormatter)
	logging.SetBackend(backendFormatterLeveled)
	if globalConfig.Debug {
		backendFormatterLeveled.SetLevel(logging.DEBUG, "hiver")
	} else {
		backendFormatterLeveled.SetLevel(logging.INFO, "hiver")
	}
}
