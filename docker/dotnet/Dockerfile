FROM --platform=$BUILDPLATFORM golang:1.26.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS dotnettools
ARG TARGETARCH
RUN dotnet tool install --tool-path /dotnet-tools dotnet-trace \
    && dotnet tool install --tool-path /dotnet-tools dotnet-gcdump \
    && dotnet tool install --tool-path /dotnet-tools dotnet-dump \
    && dotnet tool install --tool-path /dotnet-tools dotnet-counters

FROM mcr.microsoft.com/dotnet/runtime:8.0
RUN apt-get update \
    && apt-get install -y procps strace \
    && apt-get clean \
    && mkdir -p /app
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=dotnettools /dotnet-tools/ /app/
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
