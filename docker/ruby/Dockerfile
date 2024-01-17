FROM golang:1.21.3-bullseye AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent

FROM rust:1.74-bullseye AS rbspybuild
WORKDIR /
RUN git clone https://github.com/brendangregg/FlameGraph
RUN git clone --depth 1 --branch v0.18.1 https://github.com/rbspy/rbspy
WORKDIR /rbspy
RUN cargo build --release

FROM bitnami/minideb:bullseye
RUN apt-get update && apt-get install -y procps strace
RUN mkdir -p /app/FlameGraph
COPY --from=rbspybuild /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=rbspybuild /rbspy/target/release/rbspy /app/rbspy

CMD [ "/app/agent" ]
