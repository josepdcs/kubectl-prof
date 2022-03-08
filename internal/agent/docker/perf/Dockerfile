FROM golang:1.17.7-buster as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-profiling
ADD . /go/src/github.com/josepdcs/kubectl-profiling
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM alpine as builder
RUN apk add --update git
RUN git clone https://github.com/brendangregg/FlameGraph

FROM alpine
RUN apk add --update perf perl
RUN mkdir -p /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
RUN mv /usr/bin/perf /app/perf
COPY --from=builder /FlameGraph /app/FlameGraph

CMD [ "/app/agent" ]
