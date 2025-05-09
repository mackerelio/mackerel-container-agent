BIN := mackerel-container-agent
VERSION := 0.11.5

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -o build/$(BIN) ./cmd/$(BIN)/...

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
	GOOS=linux go build -o build/$(BIN) ./cmd/$(BIN)/...

.PHONY: docker
docker:
	docker build -t $(BIN) -t $(BIN):$(VERSION) --target container-agent .

.PHONY: docker-with-plugins
docker-with-plugins:
	docker build -t $(BIN) -t $(BIN):$(VERSION) --target container-agent-with-plugins .

.PHONY: version
version:
	echo $(VERSION)
