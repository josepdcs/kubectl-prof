FROM golang:1.24.5-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM alpine:3.22.1 AS builder
RUN apk add --no-cache git  \
    && git clone https://github.com/brendangregg/FlameGraph

FROM alpine:3.22.1
RUN apk add --no-cache bash bcc-tools perl procps strace  \
    && mkdir -p /app/FlameGraph
COPY --from=builder /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/bpf/profile.py /app/bcc-profiler/profile
RUN chmod +x /app/bcc-profiler/profile
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]