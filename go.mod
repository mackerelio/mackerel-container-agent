module github.com/mackerelio/mackerel-container-agent

go 1.17

require (
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/Songmu/retry v0.1.0
	github.com/Songmu/timeout v0.4.0
	github.com/Songmu/wrapcommander v0.1.0 // indirect
	github.com/aws/aws-sdk-go v1.36.0
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/containerd/cgroups v0.0.0-20190328223300-4994991857f9 // indirect
	github.com/containernetworking/cni v0.7.1 // indirect
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20190413172335-9b2eaa8a5d66
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mackerelio/go-osstat v0.0.0-20190412014440-b90c0b34edef
	github.com/mackerelio/golib v0.0.0-20190411032134-c87047ca454e
	github.com/mackerelio/mackerel-client-go v0.21.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runtime-spec v0.1.2-0.20190408193819-a1b50f621a48 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.28.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2 // indirect
	google.golang.org/grpc v1.40.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
)

require (
	github.com/aws/amazon-ecs-agent/agent v0.0.0-20220114175954-80ad4ca41f31
	k8s.io/kubelet v0.23.3
)

require (
	github.com/awslabs/go-config-generator-for-fluentd-and-fluentbit v0.0.0-20210308162251-8959c62cb8f9 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/containernetworking/plugins v0.8.6 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/vishvananda/netlink v0.0.0-20181108222139-023a6dafdcdf // indirect
	github.com/vishvananda/netns v0.0.0-20180720170159-13995c7128cc // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.6-0.20210820212750-d4cc65f0b2ff // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace github.com/etcd-io/bbolt => go.etcd.io/bbolt v1.3.5
