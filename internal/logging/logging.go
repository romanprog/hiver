package logging

import (
	"os"

	"github.com/op/go-logging"
	"github.com/romanprog/hiver/internal/config"
)

var logFormatDefault = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
var logFormatDebug = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{callpath} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// log - main logger util.
// loggingInit - initial function for logging subsystem.
func init() {
	// Create backend for logs output.
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	var logFormat logging.Formatter
	if config.Global.Debug {
		logFormat = logFormatDebug
	} else {
		logFormat = logFormatDefault
	}

	// Set log format.
	backendFormatter := logging.NewBackendFormatter(backend, logFormat)
	// Set the backends to be used and set logging level.
	backendFormatterLeveled := logging.AddModuleLevel(backendFormatter)
	logging.SetBackend(backendFormatterLeveled)
	// Set logging level.
	if config.Global.Debug {
		backendFormatterLeveled.SetLevel(logging.DEBUG, "hiver")
	} else {
		backendFormatterLeveled.SetLevel(logging.INFO, "hiver")
	}
}
