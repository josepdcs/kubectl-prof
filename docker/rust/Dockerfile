FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM rust:1.89-trixie AS rusttools
RUN cargo install flamegraph

FROM bitnami/minideb:trixie
# Install dependencies including perf and sudo
RUN mkdir -p /app && \
    apt-get update && \
    apt-get install -y procps strace linux-perf sudo binutils && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Verify perf is accessible
RUN perf --version

# Add /app to PATH so flamegraph can find perf wrapper
ENV PATH="/app:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=rusttools /usr/local/cargo/bin/flamegraph /app/flamegraph
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/perf-wrapper.sh /app/perf
RUN chmod +x /app/get-ps-command.sh && chmod +x /app/perf

CMD [ "/app/agent" ]
