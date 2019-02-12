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

// Summary ...
type Summary struct {
	Node NodeStats  `json:"node"`
	Pods []PodStats `json:"pods"`
}

// NodeStats ...
type NodeStats struct {
	NodeName         string           `json:"nodeName"`
	SystemContainers []ContainerStats `json:"systemContainers,omitempty"`
	StartTime        time.Time        `json:"startTime"`
	CPU              *CPUStats        `json:"cpu,omitempty"`
	Memory           *MemoryStats     `json:"memory,omitempty"`
	Network          *NetworkStats    `json:"network,omitempty"`
	Fs               *FsStats         `json:"fs,omitempty"`
	Runtime          *RuntimeStats    `json:"runtime,omitempty"`
}

// RuntimeStats ...
type RuntimeStats struct {
	ImageFs *FsStats `json:"imageFs,omitempty"`
}

// PodStats ...
type PodStats struct {
	PodRef           PodReference     `json:"podRef"`
	StartTime        time.Time        `json:"startTime"`
	Containers       []ContainerStats `json:"containers"`
	CPU              *CPUStats        `json:"cpu,omitempty"`
	Memory           *MemoryStats     `json:"memory,omitempty"`
	Network          *NetworkStats    `json:"network,omitempty"`
	VolumeStats      []VolumeStats    `json:"volume,omitempty"`
	EphemeralStorage *FsStats         `json:"ephemeral-storage,omitempty"`
}

// ContainerStats ...
type ContainerStats struct {
	Name               string              `json:"name"`
	StartTime          time.Time           `json:"startTime"`
	CPU                *CPUStats           `json:"cpu,omitempty"`
	Memory             *MemoryStats        `json:"memory,omitempty"`
	Accelerators       []AcceleratorStats  `json:"accelerators,omitempty"`
	Rootfs             *FsStats            `json:"rootfs,omitempty"`
	Logs               *FsStats            `json:"logs,omitempty"`
	UserDefinedMetrics []UserDefinedMetric `json:"userDefinedMetrics,omitmepty"`
}

// PodReference ...
type PodReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

// InterfaceStats ...
type InterfaceStats struct {
	Name     string  `json:"name"`
	RxBytes  *uint64 `json:"rxBytes,omitempty"`
	RxErrors *uint64 `json:"rxErrors,omitempty"`
	TxBytes  *uint64 `json:"txBytes,omitempty"`
	TxErrors *uint64 `json:"txErrors,omitempty"`
}

// NetworkStats ...
type NetworkStats struct {
	Time           time.Time `json:"time"`
	InterfaceStats `json:",inline"`
	Interfaces     []InterfaceStats `json:"interfaces,omitempty"`
}

// CPUStats ...
type CPUStats struct {
	Time                 time.Time `json:"time"`
	UsageNanoCores       *uint64   `json:"usageNanoCores,omitempty"`
	UsageCoreNanoSeconds *uint64   `json:"usageCoreNanoSeconds,omitempty"`
}

// MemoryStats ...
type MemoryStats struct {
	Time            time.Time `json:"time"`
	AvailableBytes  *uint64   `json:"availableBytes,omitempty"`
	UsageBytes      *uint64   `json:"usageBytes,omitempty"`
	WorkingSetBytes *uint64   `json:"workingSetBytes,omitempty"`
	RSSBytes        *uint64   `json:"rssBytes,omitempty"`
	PageFaults      *uint64   `json:"pageFaults,omitempty"`
	MajorPageFaults *uint64   `json:"majorPageFaults,omitempty"`
}

// AcceleratorStats ...
type AcceleratorStats struct {
	Make        string `json:"make"`
	Model       string `json:"model"`
	ID          string `json:"id"`
	MemoryTotal uint64 `json:"memoryTotal"`
	MemoryUsed  uint64 `json:"memoryUsed"`
	DutyCycle   uint64 `json:"dutyCycle"`
}

// VolumeStats ...
type VolumeStats struct {
	FsStats
	Name   string        `json:"name,omitempty"`
	PVCRef *PVCReference `json:"pvcRef,omitempty"`
}

// PVCReference ...
type PVCReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// FsStats ...
type FsStats struct {
	Time           time.Time `json:"time"`
	AvailableBytes *uint64   `json:"availableBytes,omitempty"`
	CapacityBytes  *uint64   `json:"capacityBytes,omitempty"`
	UsedBytes      *uint64   `json:"usedBytes,omitempty"`
	InodesFree     *uint64   `json:"inodesFree,omitempty"`
	Inodes         *uint64   `json:"inodes,omitempty"`
	InodesUsed     *uint64   `json:"inodesUsed,omitempty"`
}

// UserDefinedMetricDescriptor ...
type UserDefinedMetricDescriptor struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Units  string            `json:"units"`
	Labels map[string]string `json:"labels,omitempty"`
}

// UserDefinedMetric ...
type UserDefinedMetric struct {
	UserDefinedMetricDescriptor `json:",inline"`
	Time                        time.Time `json:"time"`
	Value                       float64   `json:"value"`
}
