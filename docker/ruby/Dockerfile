FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM rust:1.89-trixie AS rbspybuild
WORKDIR /
RUN git clone https://github.com/brendangregg/FlameGraph  \
    && git clone --depth 1 --branch v0.39.0 https://github.com/rbspy/rbspy
WORKDIR /rbspy
RUN cargo build --release

FROM bitnami/minideb:trixie
RUN apt-get update  \
    && apt-get install -y procps strace  \
    && apt-get clean \
    && mkdir -p /app/FlameGraph
COPY --from=rbspybuild /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=rbspybuild /rbspy/target/release/rbspy /app/rbspy
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
