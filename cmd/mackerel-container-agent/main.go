package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/mackerelio/golib/logging"

	"github.com/mackerelio/mackerel-container-agent/agent"
	"github.com/mackerelio/mackerel-container-agent/config"
)

const cmdName = "mackerel-container-agent"

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
	default:
		logging.SetLogLevel(logging.INFO)
	}

	logger.Debugf("MACKEREL_APIBASE=%s", config.MaskEnvValue(os.Getenv("MACKEREL_APIBASE")))
	logger.Debugf("MACKEREL_APIKEY=%s", config.MaskEnvValue(os.Getenv("MACKEREL_APIKEY")))

	env := []string{
		"MACKEREL_AGENT_CONFIG_POLLING_DURATION_MINUTES",
		"MACKEREL_AGENT_CONFIG",
		"MACKEREL_AGENT_PLUGIN_META",
		"MACKEREL_CONTAINER_PLATFORM",
		"MACKEREL_HOST_STATUS_ON_START",
		"MACKEREL_IGNORE_CONTAINER",
		"MACKEREL_KUBERNETES_KUBELET_HOST",
		"MACKEREL_KUBERNETES_KUBELET_INSECURE_TLS",
		"MACKEREL_KUBERNETES_KUBELET_READ_ONLY_PORT",
		"MACKEREL_KUBERNETES_NAMESPACE",
		"MACKEREL_KUBERNETES_POD_NAME",
		"MACKEREL_LOG_LEVEL",
		"MACKEREL_ROLES",
		"MACKEREL_DISPLAY_NAME",
		"MACKEREL_MEMO",
	}
	for _, v := range env {
		logger.Debugf("%s=%s", v, os.Getenv(v))
	}

	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	version, revision := fromVCS()
	logger.Infof("starting %s (version:%s, revision:%s)", cmdName, version, revision)
	if err := agent.NewAgent(version, revision).Run(args); err != nil {
		logger.Errorf("%s", err)
		return 1
	}
	return 0
}

func fromVCS() (version, rev string) {
	version = "unknown"
	rev = "unknown"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	// trim a prefix `v`
	version, _ = strings.CutPrefix(info.Main.Version, "v")

	// strings like "v0.1.2-0.20060102150405-xxxxxxxxxxxx" are long, so they are cut out.
	if strings.Contains(version, "-") {
		index := strings.IndexRune(version, '-')
		version = version[0:index]
	}

	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			// emulate "git rev-parse --short HEAD"
			rev = s.Value[0:min(len(s.Value), 7)]
			return
		}
	}
	return
}
