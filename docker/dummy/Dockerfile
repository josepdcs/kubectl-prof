FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM alpine:3.22.1
RUN apk add --no-cache bash coreutils procps strace &&  \
    mkdir -p /app
COPY --from=agentbuild /go/bin/agent /app

CMD [ "/app/agent" ]