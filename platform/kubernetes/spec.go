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
	ClusterName     string            `json:"clusterName,omitempty"`
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
func (g *specGenerator) Generate(ctx context.Context) (interface{}, error) {
	p, err := g.client.GetPod(ctx)
	if err != nil {
		return nil, err
	}

	var spec = &podSpec{
		ClusterName:    p.Metadata.ClusterName,
		Namespace:      p.Metadata.Namespace,
		Name:           p.Metadata.Name,
		UID:            p.Metadata.UID,
		ResouceVersion: p.Metadata.ResouceVersion,
		Labels:         p.Metadata.Labels,
	}
	if p.Metadata.OwnerReferences != nil {
		ownerRefs := make([]ownerReference, len(p.Metadata.OwnerReferences))
		spec.OwnerReferences = ownerRefs
		for i, r := range p.Metadata.OwnerReferences {
			ownerRefs[i] = ownerReference{
				Kind: r.Kind,
				Name: r.Name,
				UID:  r.UID,
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
				containerSpec.Resources.Limits = resourceList(limits)
			}

			if c.Ports != nil {
				ports := make([]containerPort, len(c.Ports))
				containerSpec.Ports = ports
				for j, p := range c.Ports {
					ports[j] = containerPort{
						ContainerPort: p.ContainerPort,
						HostPort:      p.HostPort,
						Name:          p.Name,
						Protocol:      p.Protocol,
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

	spec.Phase = p.Status.Phase
	spec.HostIP = p.Status.HostIP
	spec.PodIP = p.Status.PodIP
	spec.StartTime = p.Status.StartTime
	if p.Status.Conditions != nil {
		conds := make([]podCondition, len(p.Status.Conditions))
		spec.Conditions = conds
		for i, c := range p.Status.Conditions {
			conds[i] = podCondition{
				Status: c.Status,
				Type:   c.Type,
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
