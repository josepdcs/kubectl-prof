ARG JAVA_VERSION_TAG=25
FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS agentbuild
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent


FROM eclipse-temurin:$JAVA_VERSION_TAG AS tools
ARG TARGETPLATFORM
ARG ASYNCP_VERSION=4.2.1
RUN apt-get update  \
    && apt-get install -y tar wget  \
    && echo "ASYNCP_VERSION=${ASYNCP_VERSION}, TARGETPLATFORM=${TARGETPLATFORM}"  \
    && "$JAVA_HOME"/bin/jlink \
         --module-path jmods \
         --add-modules jdk.jcmd \
         --strip-debug \
         --no-man-pages \
         --no-header-files \
         --compress=2 \
         --output /jcmd; \
    if [ "$TARGETPLATFORM" = "linux/amd64" ]; then ARCHITECTURE=x64;  \
    elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then ARCHITECTURE=arm64;  \
    else ARCHITECTURE=x64;  \
    fi; \
    wget -O async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz https://github.com/async-profiler/async-profiler/releases/download/v${ASYNCP_VERSION}/async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz \
    && tar -xf async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz  \
    && mv async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE} async-profiler

FROM bitnami/minideb:trixie
RUN apt-get update  \
    && apt-get install -y gzip procps strace sudo \
    && apt-get clean \
    && mkdir -p /app/async-profiler/build
COPY --from=agentbuild /go/bin/agent /app
COPY --from=tools /async-profiler /app/async-profiler/build
COPY --from=tools /jcmd /opt/jdk
RUN mkdir -p /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/jfr-profile.jfc /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
