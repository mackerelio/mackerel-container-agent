BIN := mackerel-container-agent
VERSION := 0.5.1
REVISION := $(shell git rev-parse --short HEAD)

export GO111MODULE=on

.PHONY: all
all: clean build

BUILD_LDFLAGS := "\
	-X main.version=$(VERSION) \
	-X main.revision=$(REVISION)"

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o build/$(BIN) ./cmd/$(BIN)/...

.PHONY: test
test:
	go test -v ./...

.PHONY: lint-deps
lint-deps:
	GO111MODULE=off go get golang.org/x/lint/golint

.PHONY: lint
lint: lint-deps
	go vet ./...
	golint -set_exit_status ./...

.PHONY: clean
clean:
	rm -fr build
	go clean ./...

.PHONY: linux
linux:
	GOOS=linux go build -ldflags=$(BUILD_LDFLAGS) -o build/$(BIN) ./cmd/$(BIN)/...

.PHONY: docker
docker: linux
	docker build -t $(BIN) -t $(BIN):$(VERSION) .

.PHONY: version
version:
	echo $(VERSION)

.PHONY: update
update:
	./internal/k8s-apis/update.sh
	go get -u ./...
	go mod tidy
