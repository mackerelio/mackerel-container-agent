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

// MachineInfo ...
type MachineInfo struct {
	NumCores       int                 `json:"num_cores"`
	CPUFrequency   uint64              `json:"cpu_frequency_khz"`
	MemoryCapacity uint64              `json:"memory_capacity"`
	HugePages      []HugePagesInfo     `json:"hugepages"`
	MachineID      string              `json:"machine_id"`
	SystemUUID     string              `json:"system_uuid"`
	BootID         string              `json:"boot_id"`
	Filesystems    []FsInfo            `json:"filesystems"`
	DiskMap        map[string]DiskInfo `json:"disk_map"`
	NetworkDevices []NetInfo           `json:"network_devices"`
	Topology       []Node              `json:"topology"`
	CloudProvider  CloudProvider       `json:"cloud_provider"`
	InstanceType   InstanceType        `json:"instance_type"`
	InstanceID     InstanceID          `json:"instance_id"`
}

// HugePagesInfo ...
type HugePagesInfo struct {
	PageSize uint64 `json:"page_size"`
	NumPages uint64 `json:"num_pages"`
}

// FsInfo ...
type FsInfo struct {
	Device      string `json:"device"`
	DeviceMajor uint64 `json:"-"`
	DeviceMinor uint64 `json:"-"`
	Capacity    uint64 `json:"capacity"`
	Type        string `json:"type"`
	Inodes      uint64 `json:"inodes"`
	HasInodes   bool   `json:"has_inodes"`
}

// DiskInfo ...
type DiskInfo struct {
	Name      string `json:"name"`
	Major     uint64 `json:"major"`
	Minor     uint64 `json:"minor"`
	Size      uint64 `json:"size"`
	Scheduler string `json:"scheduler"`
}

// NetInfo ...
type NetInfo struct {
	Name       string `json:"name"`
	MacAddress string `json:"mac_address"`
	Speed      int64  `json:"speed"`
	Mtu        int64  `json:"mtu"`
}

// Node ...
type Node struct {
	ID     int     `json:"node_id"`
	Memory uint64  `json:"memory"`
	Cores  []Core  `json:"cores"`
	Caches []Cache `json:"caches"`
}

// Core ...
type Core struct {
	ID      int     `json:"core_id"`
	Threads []int   `json:"thread_ids"`
	Caches  []Cache `json:"caches"`
}

// Cache ...
type Cache struct {
	Size  uint64 `json:"size"`
	Type  string `json:"type"`
	Level int    `json:"level"`
}

// CloudProvider ...
type CloudProvider string

// InstanceType ...
type InstanceType string

// InstanceID ...
type InstanceID string
