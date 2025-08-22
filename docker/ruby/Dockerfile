FROM golang:1.24.5-bullseye AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM rust:1.89-bullseye AS rbspybuild
WORKDIR /
RUN git clone https://github.com/brendangregg/FlameGraph  \
    && git clone --depth 1 --branch v0.36.1 https://github.com/rbspy/rbspy
WORKDIR /rbspy
RUN cargo build --release

FROM bitnami/minideb:bullseye
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
