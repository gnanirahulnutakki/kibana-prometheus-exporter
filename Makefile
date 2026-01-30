.PHONY: build run test clean docker docker-push lint fmt

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOCKER_REPO ?= ghcr.io/gnanirahulnutakki/kibana-prometheus-exporter

LDFLAGS := -ldflags="-w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

build:
	@echo "Building kibana-prometheus-exporter $(VERSION)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/kibana-exporter ./cmd/exporter

run: build
	./bin/kibana-exporter --kibana-url=http://localhost:5601

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .

clean:
	rm -rf bin/

# Docker targets
docker:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_REPO):$(VERSION) \
		-t $(DOCKER_REPO):latest \
		.

docker-push: docker
	docker push $(DOCKER_REPO):$(VERSION)
	docker push $(DOCKER_REPO):latest

# Security scanning
scan:
	@echo "Scanning dependencies for vulnerabilities..."
	govulncheck ./...
	@echo "Scanning container image..."
	trivy image $(DOCKER_REPO):$(VERSION)

# Generate go.sum
deps:
	go mod tidy
	go mod verify
