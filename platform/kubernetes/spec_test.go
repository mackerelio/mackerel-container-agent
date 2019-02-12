package kubernetes

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

func TestGenerateSpec(t *testing.T) {
	client := kubelet.NewMockClient(
		kubelet.MockGetPod(func(context.Context) (*kubelet.Pod, error) {
			raw, err := ioutil.ReadFile("kubelet/testdata/pods.json")
			if err != nil {
				return nil, err
			}
			var podList kubelet.PodList
			if err := json.Unmarshal(raw, &podList); err != nil {
				return nil, err
			}
			for _, pod := range podList.Items {
				if pod.Metadata.Namespace == "default" && pod.Metadata.Name == "myapp" {
					return pod, nil
				}
			}
			return nil, nil
		}),
	)
	generator := newSpecGenerator(client)
	got, err := generator.Generate(context.Background())
	if err != nil {
		t.Errorf("should not raise error: %v", err)
	}
	v, ok := got.(*agentSpec.CloudHostname)
	if !ok {
		t.Errorf("Generate() should return *spec.CloudHostname, got %T", got)
	}

	expected := &agentSpec.CloudHostname{
		Cloud: &mackerel.Cloud{
			Provider: string(platform.Kubernetes),
			MetaData: &podSpec{
				// Metadata
				Name:            "myapp",
				UID:             "ec8c70d0-93c8-11e8-a6ea-025000000001",
				Namespace:       "default",
				ResouceVersion:  "112885",
				Labels:          map[string]string{"app": "myapp"},
				OwnerReferences: []ownerReference(nil),
				ClusterName:     "",

				// Spec
				HostNetwork: false,
				NodeName:    "docker-for-desktop",
				Containers: []container{
					container{
						Name:    "nginx",
						Image:   "nginx:alpine",
						Command: []string(nil),
						Args:    []string(nil),
						Ports: []containerPort{
							containerPort{
								ContainerPort: 80,
								HostPort:      0,
								Name:          "httpd",
								Protocol:      "TCP",
								HostIP:        "",
							},
						},
						ContainerID: "docker://651bc0955f4074659ffb96321c941a92c8da879f700e3d8f5cfb52ea4a95f4a8",
					},
					container{
						Name:    "mackerel-container-agent",
						Image:   "mackerel-container-agent:0.0.1",
						Command: []string(nil),
						Args:    []string(nil),
						Resources: resourceRequirements{
							Limits: resourceList{
								"cpu":    "250m",
								"memory": "128Mi",
							},
						},
						Ports:       []containerPort(nil),
						ContainerID: "docker://85ecaca37f3f9a9b79388bc4b6706b824fcd038259dd4d786b7ce853326d00e9",
					},
				},

				// Status
				Phase:  "Running",
				HostIP: "192.168.65.3",
				PodIP:  "10.1.0.6",
				Conditions: []podCondition{
					podCondition{
						Type:   "Initialized",
						Status: "True",
					},
					podCondition{
						Type:   "Ready",
						Status: "True",
					},
					podCondition{
						Type:   "PodScheduled",
						Status: "True",
					},
				},
				StartTime: func() *time.Time { t, _ := time.Parse(time.RFC3339Nano, "2018-07-30T07:19:40Z"); return &t }(),
			},
		},
		Hostname: "myapp",
	}

	if !reflect.DeepEqual(v, expected) {
		t.Errorf("Generate() expected %#v, got %#v", expected, v)
	}
}
