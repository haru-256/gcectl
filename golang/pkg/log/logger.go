package log

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

func getLevel() log.Level {
	level := os.Getenv("GCE_COMMANDS_LOG_LEVEL")
	if level == "" {
		level = "INFO"
	}
	switch level {
	case "INFO":
		return log.InfoLevel
	case "DEBUG":
		return log.DebugLevel
	default:
		panic(fmt.Sprintf("invalid log level: %s", level))
	}
}

var Logger = log.NewWithOptions(os.Stderr, log.Options{
	Level:           getLevel(),
	ReportCaller:    true,
	ReportTimestamp: true,
})
