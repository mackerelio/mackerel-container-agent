package main

import (
	"os"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/agent"
)

const cmdName = "mackerel-container-agent"

var version, revision string

var logger = logging.GetLogger("main")

func main() {
	logLevel := os.Getenv("MACKEREL_LOG_LEVEL")
	switch logLevel {
	case "TRACE":
		logging.SetLogLevel(logging.TRACE)
	case "DEBUG":
		logging.SetLogLevel(logging.DEBUG)
	case "INFO":
		logging.SetLogLevel(logging.INFO)
	case "WARNING":
		logging.SetLogLevel(logging.WARNING)
	case "ERROR":
		logging.SetLogLevel(logging.ERROR)
	case "CRITICAL":
		logging.SetLogLevel(logging.CRITICAL)
	}

	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	logger.Infof("starting %s (version:%s, revision:%s)", cmdName, version, revision)
	if err := agent.NewAgent(version, revision).Run(args); err != nil {
		logger.Errorf("%s", err)
		return 1
	}
	return 0
}
