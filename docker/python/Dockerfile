FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM python:3.13.6-slim-bullseye AS pyspybuild
RUN pip3 install py-spy==0.4.1  \
    && apt-get update && apt-get install -y git \
    && git clone https://github.com/brendangregg/FlameGraph

FROM python:3.13.6-alpine
RUN apk add --update --no-cache perl procps strace  \
    && mkdir -p /app/FlameGraph
COPY --from=pyspybuild /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=pyspybuild /usr/local/bin/py-spy /app/py-spy
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]