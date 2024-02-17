FROM golang:1.21.6-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent

FROM php:8.3.3-fpm-alpine AS phpspybuild
RUN apk add --update --no-cache alpine-sdk
WORKDIR /
RUN git clone --recursive https://github.com/adsr/phpspy.git
WORKDIR /phpspy
RUN make
WORKDIR /
RUN git clone https://github.com/brendangregg/FlameGraph

FROM php:7.3-fpm-alpine
RUN apk add --update --no-cache perl procps strace
RUN mkdir -p /app/FlameGraph
COPY --from=phpspybuild /phpspy/phpspy /app/phpspy
COPY --from=phpspybuild /phpspy/stackcollapse-phpspy.pl /app/stackcollapse-phpspy.pl
COPY --from=phpspybuild /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]