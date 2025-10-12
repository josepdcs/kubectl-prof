FROM golang:1.25.2-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM alpine:3.22.1
RUN apk add --no-cache bash coreutils procps strace &&  \
    mkdir -p /app
COPY --from=agentbuild /go/bin/agent /app

CMD [ "/app/agent" ]