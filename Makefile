VERSION := 0.0.1
CLI_NAME := kubectl-flame
CLI_DIR := ./cli/
BUILD_DIR := bin
REGISTRY := docker.io
DOCKER_BASE_IMAGE := josepdcs/kubectl-flame
DOCKER_JVM_IMAGE := $(DOCKER_BASE_IMAGE):$(VERSION)-jvm
DOCKERFILE_JVM := ./agent/docker/jvm/Dockerfile
DOCKER_JVM_ALPINE_IMAGE := $(DOCKER_BASE_IMAGE):$(VERSION)-jvm-alpine
DOCKERFILE_JVM_ALPINE := ./agent/docker/jvm/alpine/Dockerfile
DOCKER_BPF_IMAGE := $(DOCKER_BASE_IMAGE):$(VERSION)-bpf
DOCKERFILE_BPF := ./agent/docker/bpf/Dockerfile

all: build-cli

.PHONY: build-dep
dep: ## Get the dependencies
	@go get -v -d ./...

.PHONY: build-cli
build-cli: dep ## Build the binary file
	@go build -ldflags="-X 'github.com/josepdcs/kubectl-flame/cli/cmd/version.semver=$(VERSION)'" -o $(BUILD_DIR)/$(CLI_NAME) -v $(CLI_DIR)

.PHONY: prepare-minikube
prepare-minikube:
	@eval $(minikube -p minikube docker-env)

.PHONY: build-docker-jvm
build-docker-jvm:
	@docker build -t ${DOCKER_JVM_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM) .

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

.PHONY: debug
debug: clean
	GOARCH=amd64 GOOS=linux go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(CLI_NAME)

.PHONY: clean
clean: ## Remove previous build
	@rm -rf $(BUILD_DIR)