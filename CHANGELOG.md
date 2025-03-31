# Changelog

## 0.11.4 (2025-03-31)

* Bump the k8s group with 3 updates #439 (dependabot[bot])
* [dependabot] some libraries as group #438 (yseto)
* replace to aws-sdk-go-v2 #437 (yseto)
* Bump golang.org/x/net from 0.33.0 to 0.36.0 #433 (dependabot[bot])
* Bump golang.org/x/sync from 0.10.0 to 0.12.0 #432 (dependabot[bot])


## 0.11.3 (2025-03-03)

* added container registry GHCR #429 (yseto)


## 0.11.2 (2025-01-28)

* Bump github.com/docker/docker from 26.1.0+incompatible to 26.1.5+incompatible #414 (dependabot[bot])
* Bump golang.org/x/net from 0.23.0 to 0.33.0 #413 (dependabot[bot])
* Bump github.com/aws/aws-sdk-go from 1.51.26 to 1.54.20 #411 (dependabot[bot])
* Bump github.com/mackerelio/mackerel-client-go from 0.31.0 to 0.34.0 #410 (dependabot[bot])
* Bump docker/build-push-action from 5 to 6 #401 (dependabot[bot])
* Bump github.com/mackerelio/go-osstat from 0.2.4 to 0.2.5 #393 (dependabot[bot])
* Bump github.com/shirou/gopsutil/v3 from 3.24.3 to 3.24.5 #391 (dependabot[bot])
* Bump golangci/golangci-lint-action from 4 to 6 #382 (dependabot[bot])
* replace deprecated apt-key command on Dockerfile #379 (Arthur1)
* Bump docker/setup-buildx-action from 2 to 3 #375 (dependabot[bot])


## 0.11.1 (2024-04-23)

* Bump github.com/aws/aws-sdk-go from 1.44.58 to 1.51.26 #371 (dependabot[bot])
* Bump github.com/docker/docker from 24.0.9+incompatible to 26.1.0+incompatible #370 (dependabot[bot])
* Bump golang.org/x/net from 0.17.0 to 0.23.0 #364 (dependabot[bot])
* Update base image, some libraries. #360 (yseto)
* Bump golang.org/x/sync from 0.0.0-20220722155255-886fb9371eb4 to 0.7.0 #357 (dependabot[bot])
* Bump github.com/shirou/gopsutil/v3 from 3.24.2 to 3.24.3 #355 (dependabot[bot])
* Bump github.com/docker/docker from 20.10.23+incompatible to 24.0.9+incompatible #350 (dependabot[bot])
* Bump actions/cache from 3 to 4 #342 (dependabot[bot])
* Bump actions/checkout from 3 to 4 #341 (dependabot[bot])
* Bump docker/login-action from 2 to 3 #340 (dependabot[bot])
* Bump docker/setup-qemu-action from 2 to 3 #339 (dependabot[bot])
* Bump golangci/golangci-lint-action from 3 to 4 #338 (dependabot[bot])
* Bump github.com/mackerelio/go-osstat from 0.2.3 to 0.2.4 #299 (dependabot[bot])


## 0.11.0 (2024-04-09)

* Set `displayName`, `memo` via agent #354 (mkadokawa-idcf)


## 0.10.0 (2024-03-08)

* Bump docker/build-push-action from 3 to 5 #335 (dependabot[bot])
* Bump aws-actions/configure-aws-credentials from 1 to 4 #334 (dependabot[bot])
* Bump actions/setup-go from 3 to 5 #333 (dependabot[bot])
* Use Go 1.22 #332 (Arthur1)
* Bump golang.org/x/net from 0.2.0 to 0.17.0 #331 (dependabot[bot])
* Use shirou/gopsutil for getting CPU information #330 (Arthur1)
* Replace interface{} to any #329 (Arthur1)
* Bump github.com/mackerelio/mackerel-client-go from 0.22.0 to 0.24.0 #269 (dependabot[bot])


## 0.9.1 (2023-02-08)

* Remove last year from license header #283 (Arthur1)
* Bump github.com/docker/docker from 20.10.17+incompatible to 20.10.23+incompatible #276 (dependabot[bot])


## 0.9.0 (2022-12-14)

* fix vulnerbilities CVE-2022-27664,CVE-2022-32149 #257 (pyama86)
* Enable to debug env #246 (wafuwafu13)


## 0.8.0 (2022-11-16)

* container.memory.<name>.usage is WorkingSetBytes #253 (yseto)
* Update k8s.io/api, k8s.io/apimachinery, k8s.io/kubelet #240 (yseto)
* Enable to set LogLevel by environment variable #237 (wafuwafu13)
* Bump github.com/mackerelio/mackerel-client-go from 0.21.2 to 0.22.0 #231 (dependabot[bot])
* Improve `NewPlatform` error messages #227 (wafuwafu13)
* Bump github.com/mackerelio/go-osstat from 0.2.2 to 0.2.3 #203 (dependabot[bot])


## 0.7.2 (2022-09-14)

* enable dependabot for github-actions ecosystem #213 (lufia)
* use golangci lint #212 (lufia)
* refactor: handle errors #211 (lufia)
* refactor: replace ioutil functions #210 (lufia)
* upgrade Go: 1.17 -> 1.19 #209 (lufia)
* Bump github.com/mackerelio/mackerel-client-go from 0.21.1 to 0.21.2 #207 (dependabot[bot])


## 0.7.1 (2022-08-04)

* refactoring agent/platform.go #194 (yseto)
* [ECS Anywhere] Change to a specific name. #191 (yseto)
* use bullseye #190 (yseto)


## 0.7.0 (2022-07-27)

* add support Amazon ECS Anywhere ( experimental ) #188 (do-su-0805)


## 0.6.4 (2022-07-20)

* Bump github.com/aws/aws-sdk-go from 1.44.37 to 1.44.58 #185 (dependabot[bot])
* Bump k8s.io/kubelet from 0.24.2 to 0.24.3 #184 (dependabot[bot])
* Bump github.com/mackerelio/mackerel-client-go from 0.21.0 to 0.21.1 #177 (dependabot[bot])
* Bump github.com/docker/docker from 20.10.16+incompatible to 20.10.17+incompatible #170 (dependabot[bot])


## 0.6.3 (2022-06-22)

* Bump github.com/aws/aws-sdk-go from 1.44.27 to 1.44.37 #175 (dependabot[bot])
* Bump k8s.io/apimachinery from 0.24.1 to 0.24.2 #174 (dependabot[bot])
* Bump k8s.io/kubelet from 0.24.1 to 0.24.2 #173 (dependabot[bot])


## 0.6.2 (2022-06-08)

* Bump github.com/aws/aws-sdk-go from 1.44.21 to 1.44.27 #167 (dependabot[bot])
* update k8s.io/api, k8s.io/apimachinery, k8s.io/kubelet #162 (yseto)


## 0.6.1 (2022-05-26)

* Bump github.com/aws/aws-sdk-go from 1.43.17 to 1.44.21 #160 (dependabot[bot])
* Bump github.com/docker/docker from 20.10.13+incompatible to 20.10.16+incompatible #157 (dependabot[bot])


## 0.6.0 (2022-03-30)

* Bump github.com/mackerelio/go-osstat from 0.2.1 to 0.2.2 #137 (dependabot[bot])
* [bug-fix] add ContainerHealthStatus json marshal/unmarshal methods #136 (pyto86pri)


## 0.5.3 (2022-03-15)

* Bump github.com/docker/docker from 20.10.12+incompatible to 20.10.13+incompatible #129 (dependabot[bot])
* Bump github.com/aws/aws-sdk-go from 1.43.7 to 1.43.17 #128 (dependabot[bot])
* Bump github.com/aws/aws-sdk-go from 1.36.0 to 1.43.7 #126 (dependabot[bot])
* Bump k8s.io/kubelet from 0.23.3 to 0.23.4 #124 (dependabot[bot])
* Update github.com/docker/docker, github.com/mackerelio/go-osstat, github.com/mackerelio/golib #119 (yseto)
* Remove dependency amazon-ecs-agent #117 (yseto)


## 0.5.2 (2022-02-02)

* use k8s.io/kubelet #113 (yseto)
* fix parseConfig #107 (yseto)
* Bump github.com/mackerelio/mackerel-client-go from 0.2.0 to 0.21.0 #105 (dependabot[bot])
* Bump github.com/Songmu/timeout from 0.3.1 to 0.4.0 #104 (dependabot[bot])
* stop using Circle CI #101 (yseto)


## 0.5.1 (2022-01-14)

* container push to ECR Public #97 (yseto)
* container build and push are disabled on Circle CI #96 (yseto)
* Add a workflow to build docker images. #95 (yseto)
* Use multi-stage builds #94 (yseto)
* test on github actions workflow #93 (yseto)
* upgrade to 1.17 #88 (lufia)
* probe, spec: reduce the flakiness of tests #87 (lufia)
* chore: pin go version in build #84 (pyto86pri)
* improve S3 config loader error messages #82 (itchyny)


## 0.5.0 (2020-10-27)

* [Kubernetes] Retrieve node CPU/memory capacity from local information, not from kubelet /spec #80 (astj)
* add dependent: gobump #74 (lufia)


## 0.4.0 (2020-07-16)

* disable HTTP/2 #73 (lufia)


## 0.3.1 (2020-05-11)

* fix reloading config interrupts collecting metrics / check reports #70 (susisu)


## 0.3.0 (2020-02-07)

* Fix typo: crated_at -> created_at #64 (aereal)
* add support Amazon EKS on Amazon Fargate(BETA) #58 (yseto)
* Remove a dependency on k8s.io/kubernetes #55 (lufia)
* Add cpu spec generator #54 (tanatana)
* Add design document #53 (itchyny)


## 0.2.0 (2019-07-16)

* add mackerel-plugin-json to Docker image #51 (hayajo)
* Delay host retirement for hangup signal and config reload #49 (itchyny)
* Implement polling duration for reloading agent config #47 (itchyny)
* Fix missing region error when using S3 for config path #46 (hayajo)


## 0.1.0 (2019-06-12)

* integrate ECS platforms #43 (hayajo)


## 0.0.5 (2019-05-30)

* Improve deployment #32 #34 #36 #37 #39 (hayajo)
* Provide the plugin bundled Docker image #30 (hayajo)
* don't use HTTP_PROXY when requesting HTTP probe #29 (hayajo)
* don't use HTTP_PROXY when requesting API #28 (hayajo)


## 0.0.4 (2019-05-16)

* add build-and-push-dockerimage script for pushing Docker Image manually #26 (hayajo)
* notify interrupt signals before creating platform #25 (itchyny)
* retry request to the "/task" API #24 (hayajo)
* Improve error message #22 (hayajo)
* Use k8s packages #19 (hayajo)
* Add banner image #20 (hayajo)
* Support Task Metadata Endpoint v3 #17 (hayajo)
* Support Go Modules #18 (hayajo)


## 0.0.3 (2019-04-04)

* Improve getting TaskID #15 (hayajo)
* Fix to get subgruop(cgroup) for the new ARN #13 (hayajo)


## 0.0.2 (2019-02-25)

* check http reponse status code #4 (hayajo)
* allow insecure access to kubelet api #2 (hayajo)


## 0.0.1 (2019-02-12)

* initial release (itchyny, hayajo)
