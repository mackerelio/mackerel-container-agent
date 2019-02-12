package procfs

import (
	"os"
	"reflect"
	"testing"
)

var cgroupFixtures = []interface{}{
	fsDirectory{
		Path: "30308",
		Mode: 0755,
	},
	fsSymLink{
		To:   "30308",
		Path: "self",
	},
	fsFile{
		Path: "30308/cgroup",
		Mode: 0644,
		Content: `9:perf_event:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
8:memory:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
7:hugetlb:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
6:freezer:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
5:devices:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
4:cpuset:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
3:cpuacct:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
2:cpu:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
1:blkio:/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307
`,
	},
}

func TestCgroup(t *testing.T) {
	fs, _ := mockFilesystem(cgroupFixtures, "", "test_new_cgroup")
	defer os.RemoveAll(fs)

	proc, err := NewProc(30308, fs)
	if err != nil {
		t.Errorf("NewProc() should not raise error: %v", err)
	}

	got, err := proc.Cgroup()
	if err != nil {
		t.Errorf("Cgroup() should not raise error: %v", err)
	}

	expected := Cgroup{
		"perf_event": {
			9,
			"perf_event",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"memory": {
			8,
			"memory",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"hugetlb": {
			7,
			"hugetlb",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"freezer": {
			6,
			"freezer",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"devices": {
			5,
			"devices",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"cpuset": {
			4,
			"cpuset",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"cpuacct": {
			3,
			"cpuacct",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"cpu": {
			2,
			"cpu",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
		"blkio": {
			1,
			"blkio",
			"/ecs/ddebc715-e444-456e-a90c-29be32fe2143/16743525313bd30e86b4c2c6b71cbd09925c7a47a19e2aa1bb07b35db2363307",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Cgroup() expected %v, got %v", expected, got)
	}
}
