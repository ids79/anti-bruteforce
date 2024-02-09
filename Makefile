BIN := "./bin"
DOCKER_IMG="brutforce"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

generate:
	rm -rf internal/pb
	mkdir -p internal/pb
	
	protoc \
	    --proto_path=api/ \
	    --go_out=internal/pb \
	    --go-grpc_out=internal/pb \
	    api/*.proto

build:
	go build -v -o $(BIN)/bruteforce -ldflags "$(LDFLAGS)" ./cmd/bruteforce/
	go build -v -o $(BIN)/cli -ldflags "$(LDFLAGS)" ./cmd/cli/

run: build
	$(BIN)/bruteforce -config ./configs/config.toml

run-c:
	$(BIN)/cli -config ./configs/config.toml	

build-img:
	docker build \
	--build-arg=LDFLAGS="$(LDFLAGS)" \
	-t $(DOCKER_IMG):1.0.1 \
	-f build/Dockerfile .

run-img: build-img
	docker run --name $(DOCKER_IMG) \
	--link pg:pg \
	-p 8081:8081 \
	-p 5051:5051 \
	$(DOCKER_IMG):1.0.1
		
up: docker-compose -f ./docker-compose.yml up --build -d

down: docker-compose down -f docker-compose.yml	

version: build
	$(BIN)/brutforce version
test:
	go test -race ./internal/... ./cmd/...

integation-test:
	go test -race ./tests/...

integat-docker-test:
	docker compose -f docker-compose.yml -f docker-compose.test.yml up --exit-code-from tester
	docker-compose down
	
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

install-goose:
	(which goose > /dev/null) || go install github.com/pressly/goose/v3/cmd/goose@latest

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint