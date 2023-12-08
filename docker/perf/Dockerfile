FROM golang:1.21.3-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent

FROM alpine:3.18.4 AS tools
RUN apk add --no-cache git
RUN git clone https://github.com/brendangregg/FlameGraph

FROM alpine:3.18.4
RUN apk add --no-cache perf perl procps
RUN mkdir -p /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
RUN mv /usr/bin/perf /app/perf
COPY --from=tools /FlameGraph /app/FlameGraph

CMD [ "/app/agent" ]
