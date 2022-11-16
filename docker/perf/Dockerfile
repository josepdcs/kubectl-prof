FROM golang:1.19.2-bullseye as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM alpine as builder
RUN apk add --update git
RUN git clone https://github.com/brendangregg/FlameGraph

FROM alpine
RUN apk add --update perf perl procps
RUN mkdir -p /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
RUN mv /usr/bin/perf /app/perf
COPY --from=builder /FlameGraph /app/FlameGraph

CMD [ "/app/agent" ]
