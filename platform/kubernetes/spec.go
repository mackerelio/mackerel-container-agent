package kubernetes

import (
	"context"
	"time"

	mackerel "github.com/mackerelio/mackerel-client-go"

	"github.com/mackerelio/mackerel-container-agent/platform"
	"github.com/mackerelio/mackerel-container-agent/platform/kubernetes/kubelet"
	agentSpec "github.com/mackerelio/mackerel-container-agent/spec"
)

type podSpec struct {
	// Metadata
	Namespace       string            `json:"namespace,omitempty"`
	Name            string            `json:"name,omitempty"`
	UID             string            `json:"uid,omitempty"`
	ResouceVersion  string            `json:"resourceVersion,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	OwnerReferences []ownerReference  `json:"ownerReferences,omitempty"`

	// Spec
	HostNetwork bool        `json:"hostNetwork,omitempty"`
	NodeName    string      `json:"nodeName,omitempty"`
	Containers  []container `json:"containers,omitempty"`

	// Status
	Phase      string         `json:"phase,omitempty"`
	HostIP     string         `json:"hostIP,omitempty"`
	PodIP      string         `json:"podIP,omitempty"`
	Conditions []podCondition `json:"conditions,omitempty"`
	StartTime  *time.Time     `json:"startTime,omitempty"`
}

type ownerReference struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
	UID  string `json:"uid,omitempty"`
}

type podCondition struct {
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

type container struct {
	// Spec
	Name      string               `json:"name"`
	Image     string               `json:"image,omitempty"`
	Command   []string             `json:"command,omitempty"`
	Args      []string             `json:"args,omitempty"`
	Resources resourceRequirements `json:"resources,omitempty"`
	Ports     []containerPort      `json:"ports,omitempty"`

	// Status
	ContainerID string `json:"containerID,omitempty"`
}

type resourceRequirements struct {
	Limits resourceList `json:"limits,omitempty"`
	// Requests ResourceList `json:"requests,omitempty"`
}

type resourceList map[string]string

type containerPort struct {
	ContainerPort int    `json:"containerport"`
	HostPort      int    `json:"hostport"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"hostip,omitempty"`
}

type specGenerator struct {
	client kubelet.Client
}

func newSpecGenerator(client kubelet.Client) *specGenerator {
	return &specGenerator{
		client: client,
	}
}
func (g *specGenerator) Generate(ctx context.Context) (any, error) {
	p, err := g.client.GetPod(ctx)
	if err != nil {
		return nil, err
	}

	var spec = &podSpec{
		Namespace:      p.Namespace,
		Name:           p.Name,
		UID:            string(p.UID),
		ResouceVersion: p.ResourceVersion,
		Labels:         p.Labels,
	}
	if p.OwnerReferences != nil {
		ownerRefs := make([]ownerReference, len(p.OwnerReferences))
		spec.OwnerReferences = ownerRefs
		for i, r := range p.OwnerReferences {
			ownerRefs[i] = ownerReference{
				Kind: r.Kind,
				Name: r.Name,
				UID:  string(r.UID),
			}
		}
	}

	spec.HostNetwork = p.Spec.HostNetwork
	spec.NodeName = p.Spec.NodeName
	if p.Spec.Containers != nil {
		containers := make([]container, len(p.Spec.Containers))
		spec.Containers = containers

		for i, c := range p.Spec.Containers {
			containerSpec := container{
				Name:    c.Name,
				Image:   c.Image,
				Command: c.Command,
				Args:    c.Args,
			}

			if limits := c.Resources.Limits; limits != nil {
				rl := resourceList{}
				for k, v := range limits {
					rl[string(k)] = v.String()
				}
				containerSpec.Resources.Limits = rl
			}

			if c.Ports != nil {
				ports := make([]containerPort, len(c.Ports))
				containerSpec.Ports = ports
				for j, p := range c.Ports {
					ports[j] = containerPort{
						ContainerPort: int(p.ContainerPort),
						HostPort:      int(p.HostPort),
						Name:          p.Name,
						Protocol:      string(p.Protocol),
						HostIP:        p.HostIP,
					}
				}
			}

			for _, cs := range p.Status.ContainerStatuses {
				if cs.Name == c.Name {
					containerSpec.ContainerID = cs.ContainerID
					break
				}
			}

			containers[i] = containerSpec
		}
	}

	spec.Phase = string(p.Status.Phase)
	spec.HostIP = p.Status.HostIP
	spec.PodIP = p.Status.PodIP
	spec.StartTime = &p.Status.StartTime.Time
	if p.Status.Conditions != nil {
		conds := make([]podCondition, len(p.Status.Conditions))
		spec.Conditions = conds
		for i, c := range p.Status.Conditions {
			conds[i] = podCondition{
				Status: string(c.Status),
				Type:   string(c.Type),
			}
		}
	}

	return &agentSpec.CloudHostname{
		Hostname: spec.Name,
		Cloud: &mackerel.Cloud{
			Provider: string(platform.Kubernetes),
			MetaData: spec,
		},
	}, nil
}
