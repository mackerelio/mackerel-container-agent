BIN := mackerel-container-agent

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
	docker build -t $(BIN) -t $(BIN):local --target container-agent .

.PHONY: docker-with-plugins
docker-with-plugins:
	docker build -t $(BIN) -t $(BIN):local --target container-agent-with-plugins .
