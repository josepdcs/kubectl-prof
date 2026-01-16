FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM alpine:3.22.1 AS builder
RUN apk add --no-cache git  \
    && git clone https://github.com/brendangregg/FlameGraph

FROM alpine:3.22.1 AS libbpf-builder
# Install build dependencies for libbpf-tools
RUN apk add --no-cache \
    build-base \
    clang \
    elfutils-dev \
    linux-headers \
    libbpf-dev \
    zlib-dev \
    git \
    bpftool

# Clone and build libbpf-tools profile
RUN git clone --depth 1 --recurse-submodules --shallow-submodules https://github.com/iovisor/bcc.git /bcc && \
    cd /bcc/libbpf-tools && \
    make profile

FROM alpine:3.22.1
# Install only runtime dependencies (no bcc-tools needed for CO-RE)
RUN apk add --no-cache bash perl procps strace libbpf \
    && mkdir -p /app/FlameGraph /app/libbpf-profiler
COPY --from=builder /FlameGraph /app/FlameGraph
COPY --from=libbpf-builder /bcc/libbpf-tools/profile /app/libbpf-profiler/profile
RUN chmod +x /app/libbpf-profiler/profile
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
