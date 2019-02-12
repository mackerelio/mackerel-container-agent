/*
Portions are:
copyright 2015 the kubernetes authors.

licensed under the apache license, version 2.0 (the "license");
you may not use this file except in compliance with the license.
you may obtain a copy of the license at

    http://www.apache.org/licenses/license-2.0

unless required by applicable law or agreed to in writing, software
distributed under the license is distributed on an "as is" basis,
without warranties or conditions of any kind, either express or implied.
see the license for the specific language governing permissions and
limitations under the license.
*/

package kubelet

import "time"

// PodList ...
type PodList struct {
	Items []*Pod `json:"items,omitempty"`
}

// Pod ...
type Pod struct {
	Spec     PodSpec     `json:"spec,omitempty"`
	Status   PodStatus   `json:"status,omitempty"`
	Metadata PodMetadata `json:"metadata,omitempty"`
}

// PodMetadata ...
type PodMetadata struct {
	Name            string            `json:"name,omitempty"`
	UID             string            `json:"uid,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	ResouceVersion  string            `json:"resourceVersion,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	OwnerReferences []OwnerReference  `json:"ownerReferences,omitempty"`
	ClusterName     string            `json:"clusterName,omitempty"`
}

// OwnerReference ...
type OwnerReference struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
	UID  string `json:"uid,omitempty"`
}

// PodSpec ...
type PodSpec struct {
	HostNetwork bool        `json:"hostNetwork,omitempty"`
	NodeName    string      `json:"nodeName,omitempty"`
	Containers  []Container `json:"containers,omitempty"`
}

// Container ...
type Container struct {
	Name      string               `json:"name"`
	Image     string               `json:"image,omitempty"`
	Command   []string             `json:"command,omitempty"`
	Args      []string             `json:"args,omitempty"`
	Resources ResourceRequirements `json:"resources,omitempty"`
	Ports     []ContainerPort      `json:"ports,omitempty"`
}

// ResourceRequirements ...
type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty"`
	Requests ResourceList `json:"requests,omitempty"`
}

// ResourceList ...
type ResourceList map[string]string

// ContainerPort ...
type ContainerPort struct {
	ContainerPort int    `json:"containerPort"`
	HostPort      int    `json:"hostPort"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"hostIP,omitempty"`
}

// PodStatus ...
type PodStatus struct {
	Phase             string            `json:"phase,omitempty"`
	HostIP            string            `json:"hostIP,omitempty"`
	PodIP             string            `json:"podIP,omitempty"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
	Conditions        []PodCondition    `json:"conditions,omitempty"`
	StartTime         *time.Time        `json:"startTime,omitempty"`
}

// PodCondition ...
type PodCondition struct {
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

// ContainerStatus ...
type ContainerStatus struct {
	Name        string `json:"name,omitempty"`
	Image       string `json:"image,omitempty"`
	ContainerID string `json:"containerID,omitempty"`
}
