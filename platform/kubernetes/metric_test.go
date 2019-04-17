package kubernetes

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	cadvisorTypes "github.com/google/cadvisor/info/v1"
	kubernetesTypes "k8s.io/api/core/v1"
	kubeletTypes "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"

	"github.com/mackerelio/mackerel-container-agent/metric"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
)

func TestGenerateStats(t *testing.T) {
	ctx := context.Background()
	client := kubelet.NewMockClient(
		kubelet.MockGetPod(func(context.Context) (*kubernetesTypes.Pod, error) {
			raw, err := ioutil.ReadFile("kubelet/testdata/pods.json")
			if err != nil {
				return nil, err
			}
			var podList kubernetesTypes.PodList
			if err := json.Unmarshal(raw, &podList); err != nil {
				return nil, err
			}
			for _, pod := range podList.Items {
				if pod.ObjectMeta.Namespace == "default" && pod.ObjectMeta.Name == "myapp" {
					return &pod, nil
				}
			}
			return nil, nil
		}),
		kubelet.MockGetPodStats(func(context.Context) (*kubeletTypes.PodStats, error) {
			raw, err := ioutil.ReadFile("kubelet/testdata/summary.json")
			if err != nil {
				return nil, err
			}
			var summary kubeletTypes.Summary
			if err := json.Unmarshal(raw, &summary); err != nil {
				return nil, err
			}
			for _, pod := range summary.Pods {
				if pod.PodRef.Namespace == "default" && pod.PodRef.Name == "myapp" {
					return &pod, nil
				}
			}
			return nil, nil
		}),
		kubelet.MockGetSpec(func(context.Context) (*cadvisorTypes.MachineInfo, error) {
			raw, err := ioutil.ReadFile("kubelet/testdata/spec.json")
			if err != nil {
				return nil, err
			}
			var info cadvisorTypes.MachineInfo
			if err := json.Unmarshal(raw, &info); err != nil {
				return nil, err
			}
			return &info, nil
		}),
	)
	generator := newMetricGenerator(client)
	_, err := generator.Generate(ctx) // Store metrics to generator.prevStats.
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}
	got, err := generator.Generate(ctx)
	if err != nil {
		t.Errorf("Generate() should not raise error: %v", err)
	}
	expected := metric.Values{
		"container.cpu.mackerel-container-agent.usage":    0.0, // Rsult is 0 because use the same data.
		"container.cpu.nginx.usage":                       0.0, // Rsult is 0 because use the same data.
		"container.cpu.mackerel-container-agent.limit":    25.0,
		"container.cpu.nginx.limit":                       200.0,
		"container.memory.mackerel-container-agent.usage": 2.6529792e+07,
		"container.memory.nginx.usage":                    1.949696e+06,
		"container.memory.mackerel-container-agent.limit": 134217728.0, // 128MiB
		"container.memory.nginx.limit":                    2096058368.0,
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Generate() expected %v, got %v", expected, got)
	}
}

func TestGetMemoryLimit(t *testing.T) {
	hostMemTotal := 2096058368.0
	name := "dummy"
	tests := []struct {
		limit    kubelet.ResourceList
		expected float64
	}{
		{
			nil,
			hostMemTotal,
		},
		{
			kubelet.ResourceList{},
			hostMemTotal,
		},
		{
			kubelet.ResourceList{"memory": "134217728"},
			134217728.0,
		},
		{
			kubelet.ResourceList{"memory": "128e6"},
			128000000.0,
		},
		{
			kubelet.ResourceList{"memory": "128M"},
			128000000.0,
		},
		{
			kubelet.ResourceList{"memory": "128Mi"},
			134217728.0,
		},
		{
			kubelet.ResourceList{"memory": "1G"},
			1000000000.0,
		},
		{
			kubelet.ResourceList{"memory": "1Gi"},
			1073741824.0,
		},
	}
	g := &metricGenerator{
		hostMemTotal: &hostMemTotal,
	}
	for _, tc := range tests {
		container := kubernetesTypes.Container{
			Name: name,
			Resources: kubernetesTypes.ResourceRequirements{
				Limits: tc.limit,
			},
		}
		got := g.getMermoryLimit(&container)
		if got != tc.expected {
			t.Errorf("getMermoryLimit() expected %.1f, got %.1f", tc.expected, got)
		}
	}
}
