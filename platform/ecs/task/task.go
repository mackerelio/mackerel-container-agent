package task

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/mackerelio/go-osstat/memory"

	"github.com/mackerelio/mackerel-container-agent/platform/ecs/agent"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/cgroupfs"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/docker"
	"github.com/mackerelio/mackerel-container-agent/platform/ecs/procfs"
)

const (
	taskArnPrefix   = "task/"
	ecsSubgroupName = "ecs"
	memorySubsystem = "memory"
)

// Task interface gets task metric values and metadata
type Task interface {
	Metadata(context.Context) (*Metadata, error)
	Stats(context.Context) (map[string]*Stats, error)
	IsPrivateNetworkMode() bool
}

type task struct {
	id                   string
	dockerID             string
	proc                 procfs.Proc
	dockerClient         docker.Client
	agentClient          agent.Client
	cgroup               cgroupfs.Cgroup
	isPrivateNetworkMode bool
	ignoreContainer      *regexp.Regexp
}

// NewTaskWithProc creates a new Task
func NewTaskWithProc(ctx context.Context, proc procfs.Proc, dockerClient docker.Client, agentClient agent.Client, cgroup cgroupfs.Cgroup, ignoreContainer *regexp.Regexp) (Task, error) {
	dockerID, err := getDockerID(proc)
	if err != nil {
		return nil, err
	}

	meta, err := agentClient.GetTaskMetadataWithDockerID(ctx, dockerID)
	if err != nil {
		return nil, err
	}

	taskID, err := parseTaskArn(meta.Arn)
	if err != nil {
		return nil, err
	}

	container, err := dockerClient.InspectContainer(ctx, dockerID)
	if err != nil {
		return nil, err
	}
	isPrivateNWM := container.HostConfig.NetworkMode.IsPrivate()

	return &task{
		id:                   taskID,
		dockerID:             dockerID,
		proc:                 proc,
		dockerClient:         dockerClient,
		agentClient:          agentClient,
		cgroup:               cgroup,
		isPrivateNetworkMode: isPrivateNWM,
		ignoreContainer:      ignoreContainer,
	}, nil
}

// Metadata gets metadata
func (t *task) Metadata(ctx context.Context) (*Metadata, error) {
	taskRes, err := t.agentClient.GetTaskMetadataWithDockerID(ctx, t.dockerID)
	if err != nil {
		return nil, err
	}
	instanceMeta, err := t.agentClient.GetInstanceMetadata(ctx)
	if err != nil {
		return nil, err
	}
	limits, err := t.getResourceLimits()
	if err != nil {
		return nil, err
	}

	var meta = &Metadata{
		// TaskResponse: taskRes
		Arn:           taskRes.Arn,
		DesiredStatus: taskRes.DesiredStatus,
		KnownStatus:   taskRes.KnownStatus,
		Family:        taskRes.Family,
		Version:       taskRes.Version,
	}

	containers := make([]Container, 0, len(taskRes.Containers))
	for _, c := range taskRes.Containers {
		if t.ignoreContainer != nil && t.ignoreContainer.MatchString(c.Name) {
			continue
		}
		containers = append(containers, Container{
			DockerID:   c.DockerID,
			DockerName: c.DockerName,
			Name:       c.Name,
		})
	}
	meta.Containers = containers

	meta.Instance = instanceMeta
	meta.Limits = limits

	return meta, nil
}

// Stats gets metric values
func (t *task) Stats(ctx context.Context) (map[string]*Stats, error) {
	meta, err := t.Metadata(ctx)
	if err != nil {
		return nil, err
	}

	var stats = make(map[string]*Stats)
	for _, c := range meta.Containers {
		sts, err := t.dockerClient.GetContainerStats(ctx, c.DockerID)
		if err != nil {
			return nil, err
		}
		cnt, err := t.dockerClient.InspectContainer(ctx, c.DockerID)
		if err != nil {
			return nil, err
		}

		var cpuLimit float64
		if cnt.HostConfig.CPUPeriod > 0 {
			cpuLimit = float64(cnt.HostConfig.CPUQuota) / float64(cnt.HostConfig.CPUPeriod)
		}

		var network = make(map[string]NetworkStats, len(sts.Networks))
		for name, s := range sts.Networks {
			network[name] = NetworkStats{
				RxBytes: s.RxBytes,
				TxBytes: s.TxBytes,
			}
		}

		stats[c.Name] = &Stats{
			CPU: CPUStats{
				Total: sts.CPUStats.CPUUsage.TotalUsage,
				Limit: cpuLimit,
			},
			Memory: MemoryStats{
				Limit: sts.MemoryStats.Limit,
				Usage: sts.MemoryStats.Usage,
				Stats: sts.MemoryStats.Stats,
			},
			Network: network,
		}
	}

	return stats, nil
}

// IsPrivateNetworkMode reports t is private network mode(default, bridge, none)
func (t *task) IsPrivateNetworkMode() bool {
	return t.isPrivateNetworkMode
}

func getDockerID(proc procfs.Proc) (string, error) {
	cgroup, err := proc.Cgroup()
	if err != nil {
		return "", err
	}
	memCgroup := cgroup[memorySubsystem]
	if memCgroup == nil {
		return "", errors.New("memory cgroup not exists")
	}
	parts := strings.Split(memCgroup.CgroupPath, string(os.PathSeparator))
	if len(parts) < 4 || parts[1] != "ecs" { // expect ["", "ecs", "task-id", "docker-id"] or ["", "ecs", "cluster-name", "task-id", "docker-id"]
		return "", fmt.Errorf("failed to parse %s", memCgroup.CgroupPath)
	}
	return parts[len(parts)-1], nil
}

func parseTaskArn(taskArn string) (string, error) {
	a, err := arn.Parse(taskArn)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(a.Resource, taskArnPrefix) { // expect "task/0d673060-1454-4039-bebd-02b97830e880"
		return "", fmt.Errorf("failed to parse %s", a.Resource)
	}
	return a.Resource[strings.LastIndexByte(a.Resource, '/')+1:], nil
}

func (t *task) getResourceLimits() (ResourceLimits, error) {
	var limits = ResourceLimits{}

	sg, err := t.getTaskSubgroup()
	if err != nil {
		return limits, err
	}

	cgCPU, err := t.cgroup.CPU(sg)
	if err != nil {
		return limits, err
	}
	if cgCPU.CfsQuotaUs == -1 {
		limits.CPU = float64(runtime.NumCPU()) * 100.0
	} else {
		limits.CPU = float64(cgCPU.CfsQuotaUs) / float64(cgCPU.CfsPeriodUs) * 100.0
	}

	cgMemory, err := t.cgroup.Memory(sg)
	if err != nil {
		return limits, err
	}
	if cgMemory.Limit == uint64(math.MaxInt64/os.Getpagesize()*os.Getpagesize()) { // Largest positive signed interger rounded down to PAGE_SIZE.
		m, err := memory.Get()
		if err != nil {
			return limits, err
		}
		limits.Memory = m.Total
	} else {
		limits.Memory = cgMemory.Limit
	}

	return limits, nil
}

func (t *task) getTaskSubgroup() (string, error) {
	cg, err := t.proc.Cgroup()
	if err != nil {
		return "", err
	}

	memSs, ok := cg[memorySubsystem]
	if !ok {
		return "", fmt.Errorf("%s subsystem not exists", memorySubsystem)
	}

	// expect "/ecs/TASK_ID" or "/ecs/CLUSTER_NAME/TASK_ID"
	return path.Dir(memSs.CgroupPath), nil
}
