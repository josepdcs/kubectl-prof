FROM golang:1.25.2-trixie AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM rust:1.89-trixie AS rusttools
RUN cargo install flamegraph \
    && test -x /usr/local/cargo/bin/flamegraph \
    && /usr/local/cargo/bin/flamegraph --version

FROM bitnami/minideb:trixie
RUN apt-get update  \
    && apt-get install -y procps strace  \
    && apt-get clean
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=rusttools /usr/local/cargo/bin/flamegraph /app/flamegraph
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
