package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
	"github.com/mackerelio/mackerel-container-agent/platform/none"
)

// NewPlatform creates a new container platform
func NewPlatform(ctx context.Context, ignoreContainer *regexp.Regexp) (platform.Platform, error) {
	p := os.Getenv("MACKEREL_CONTAINER_PLATFORM")

	switch platform.Type(p) {

	case platform.ECSAwsvpc, platform.ECSv3:
		logger.Warningf("%q platform is deprecated. Please use %q platform", p, platform.ECS)
		fallthrough

	case platform.ECS, platform.Fargate:
		metadataURI, err := getEnvValue("ECS_CONTAINER_METADATA_URI")
		if err != nil {
			return nil, err
		}
		executionEnv, err := getEnvValue("AWS_EXECUTION_ENV")
		if err != nil {
			return nil, err
		}
		return ecs.NewECSPlatform(ctx, metadataURI, executionEnv, ignoreContainer)

	case platform.Kubernetes:
		useReadOnlyPort := true
		insecureTLS := false
		serviceHost, err := getEnvValue("KUBERNETES_SERVICE_HOST")
		if err != nil {
			return nil, err
		}
		servicePort, err := getEnvValue("KUBERNETES_SERVICE_PORT")
		if err != nil {
			servicePort = "443"
		}
		kubeletHost, err := getEnvValue("MACKEREL_KUBERNETES_KUBELET_HOST")
		if err != nil {
			return nil, err
		}
		kubeletPort, err := getEnvValue("MACKEREL_KUBERNETES_KUBELET_READ_ONLY_PORT")
		if err != nil {
			kubeletPort = kubelet.DefaultReadOnlyPort
		}
		if kubeletPort == "0" {
			useReadOnlyPort = false
			kubeletPort, err = getEnvValue("MACKEREL_KUBERNETES_KUBELET_PORT")
			if err != nil {
				kubeletPort = kubelet.DefaultPort
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
		nodeName, err := getEnvValue("MACKEREL_KUBERNETES_NODE_NAME")
		if err != nil {
			return nil, err
		}
		return kubernetes.NewKubernetesPlatform(serviceHost, servicePort, kubeletHost, kubeletPort, useReadOnlyPort, insecureTLS, namespace, podName, nodeName, ignoreContainer)

	case platform.EKSOnFargate:
		serviceHost, err := getEnvValue("KUBERNETES_SERVICE_HOST")
		if err != nil {
			return nil, err
		}
		servicePort, err := getEnvValue("KUBERNETES_SERVICE_PORT")
		if err != nil {
			servicePort = "443"
		}
		namespace, err := getEnvValue("MACKEREL_KUBERNETES_NAMESPACE")
		if err != nil {
			return nil, err
		}
		podName, err := getEnvValue("MACKEREL_KUBERNETES_POD_NAME")
		if err != nil {
			return nil, err
		}
		nodeName, err := getEnvValue("MACKEREL_KUBERNETES_NODE_NAME")
		if err != nil {
			return nil, err
		}
		return kubernetes.NewEKSOnFargatePlatform(serviceHost, servicePort, namespace, podName, nodeName, ignoreContainer)

	// for testing & debugging on local machine
	case platform.None:
		return none.NewNonePlatform()

	default:
		return nil, errors.New("platform not specified")
	}
}

func getEnvValue(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return value, fmt.Errorf("%s environment variable is not set", name)
	}
	return value, nil
}
