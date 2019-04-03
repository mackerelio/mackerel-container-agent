package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs"
	ecsInstance "github.com/mackerelio/mackerel-container-agent/platform/ecs/instance"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsawsvpc"
	"github.com/mackerelio/mackerel-container-agent/platform/ecsv3"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
)

// NewPlatform creates a new container platform
func NewPlatform(ctx context.Context, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	switch platform.Type(os.Getenv("MACKEREL_CONTAINER_PLATFORM")) {

	case platform.ECS:
		instanceClient, err := ecsInstance.NewClient(ecsInstance.DefaultURL)
		if err != nil {
			return nil, err
		}
		return ecs.NewECSPlatform(ctx, instanceClient, ignoreContainer)

	case platform.ECSAwsvpc:
		return ecsawsvpc.NewECSAwsvpcPlatform(false, ignoreContainer)

	case platform.ECSv3:
		metadataURI := os.Getenv("ECS_CONTAINER_METADATA_URI")
		return ecsv3.NewECSPlatform(metadataURI, false, ignoreContainer)

	case platform.Fargate:
		return ecsawsvpc.NewECSAwsvpcPlatform(true, ignoreContainer)

	case platform.Kubernetes:
		useReadOnlyPort := true
		insecureTLS := false
		host, err := getEnvValue("MACKEREL_KUBERNETES_KUBELET_HOST")
		if err != nil {
			return nil, err
		}
		port, err := getEnvValue("MACKEREL_KUBERNETES_KUBELET_READ_ONLY_PORT")
		if err != nil {
			port = kubelet.DefaultReadOnlyPort
		}
		if port == "0" {
			useReadOnlyPort = false
			port, err = getEnvValue("MACKEREL_KUBERNETES_KUBELET_PORT")
			if err != nil {
				port = kubelet.DefaultPort
			}
			_, err := getEnvValue("MACKEREL_KUBERNETES_KUBELET_INSECURE_TLS")
			if err == nil {
				insecureTLS = true
			}
		}
		namespace, err := getEnvValue("MACKEREL_KUBERNETES_NAMESPACE")
		if err != nil {
			return nil, err
		}
		podName, err := getEnvValue("MACKEREL_KUBERNETES_POD_NAME")
		if err != nil {
			return nil, err
		}
		return kubernetes.NewKubernetesPlatform(host, port, useReadOnlyPort, insecureTLS, namespace, podName, ignoreContainer)

	default:
		return nil, errors.New("platform not specified")
	}
}

func getEnvValue(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return value, fmt.Errorf("please set the %s environment variable", name)
	}
	return value, nil
}
