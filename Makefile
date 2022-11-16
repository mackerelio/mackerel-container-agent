BIN := mackerel-container-agent
VERSION := 0.8.0
REVISION := $(shell git rev-parse --short HEAD)

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

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -fr build
	go clean ./...

.PHONY: linux
linux:
	GOOS=linux go build -ldflags=$(BUILD_LDFLAGS) -o build/$(BIN) ./cmd/$(BIN)/...

.PHONY: docker
docker:
	docker build -t $(BIN) -t $(BIN):$(VERSION) --target container-agent .

.PHONY: version
version:
	echo $(VERSION)
