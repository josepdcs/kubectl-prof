FROM golang:1.18.1-buster as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM rust:1.50 AS rbspybuild
WORKDIR /
RUN git clone --depth 1 --branch v0.8.1 https://github.com/rbspy/rbspy
RUN cd rbspy && cargo build --release

FROM bitnami/minideb:stretch
RUN apt-get update && apt-get install -y procps
RUN mkdir /app
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=rbspybuild /rbspy/target/release/rbspy /app/rbspy

CMD [ "/app/agent" ]
