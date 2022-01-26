# Changelog

## 0.5.2 (2022-01-26)

* fix parseConfig #107 (yseto)
* Bump github.com/mackerelio/mackerel-client-go from 0.2.0 to 0.21.0 #105 (dependabot[bot])
* added dependabot.yml #102 (yseto)
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
