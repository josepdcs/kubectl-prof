VERSION ?= 0.2.0-dev
CLI_NAME ?= kubectl-perf
CLI_DIR ?= ./cmd/cli/
BUILD_DIR ?= bin
REGISTRY ?= docker.io
DOCKER_BASE_IMAGE ?= josepdcs/kubectl-perf
DOCKER_JVM_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-jvm
DOCKERFILE_JVM ?= ./pkg/agent/docker/jvm/Dockerfile
DOCKER_JVM_ALPINE_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-jvm-alpine
DOCKERFILE_JVM_ALPINE ?= ./pkg/agent/docker/jvm/alpine/Dockerfile
DOCKER_BPF_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-bpf
DOCKERFILE_BPF ?= ./pkg/agent/docker/bpf/Dockerfile
DOCKER_PERF_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-perf
DOCKERFILE_PERF ?= ./pkg/agent/docker/perf/Dockerfile
DOCKER_PYTHON_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-python
DOCKERFILE_PYTHON ?= ./pkg/agent/docker/python/Dockerfile
DOCKER_RUBY_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-ruby
DOCKERFILE_RUBY ?= ./pkg/agent/docker/ruby/Dockerfile

all: build-cli push-docker-jvm push-docker-jvm-alpine push-docker-bpf push-docker-perf push-docker-python push-docker-ruby

.PHONY: build-dep
dep: ## Get the dependencies
	@go get -v -d ./...

.PHONY: build-cli
build-cli: dep ## Build the binary file
	@go build -ldflags="-X 'github.com/josepdcs/kubectl-perf/pkg/cli/version.semver=$(VERSION)'" -o $(BUILD_DIR)/$(CLI_NAME) -v $(CLI_DIR)

.PHONY: build-docker-jvm
build-docker-jvm:
	@docker build --no-cache -t ${DOCKER_JVM_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM) .

.PHONY: push-docker-jvm
push-docker-jvm: build-docker-jvm
	@docker push $(REGISTRY)/$(DOCKER_JVM_IMAGE)

.PHONY: build-docker-jvm-alpine
build-docker-jvm-alpine:
	@docker build -t ${DOCKER_JVM_ALPINE_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM_ALPINE) .

.PHONY: push-docker-jvm-alpine
push-docker-jvm-alpine: build-docker-jvm-alpine
	@docker push $(REGISTRY)/$(DOCKER_JVM_ALPINE_IMAGE)

.PHONY: build-docker-bpf
build-docker-bpf:
	docker build -t ${DOCKER_BPF_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_BPF) .

.PHONY: push-docker-bpf
push-docker-bpf: build-docker-bpf
	@docker push $(REGISTRY)/$(DOCKER_BPF_IMAGE)

.PHONY: build-docker-perf
build-docker-perf:
	docker build -t ${DOCKER_PERF_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_PERF) .

.PHONY: push-docker-perf
push-docker-perf: build-docker-perf
	@docker push $(REGISTRY)/$(DOCKER_PERF_IMAGE)

.PHONY: build-docker-python
build-docker-python:
	docker build -t ${DOCKER_PYTHON_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_PYTHON) .

.PHONY: push-docker-python
push-docker-python: build-docker-python
	@docker push $(REGISTRY)/$(DOCKER_PYTHON_IMAGE)

.PHONY: build-docker-ruby
build-docker-ruby:
	docker build -t ${DOCKER_RUBY_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_RUBY) .

.PHONY: push-docker-ruby
push-docker-ruby: build-docker-ruby
	@docker push $(REGISTRY)/$(DOCKER_RUBY_IMAGE)

.PHONY: debug
debug: clean
	GOARCH=amd64 GOOS=linux go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(CLI_NAME)

.PHONY: clean
clean: ## Remove previous build
	@rm -rf $(BUILD_DIR)