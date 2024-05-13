ARG JAVA_VERSION_TAG=21
ARG ASYNCP_VERSION=2.9
FROM --platform=$BUILDPLATFORM golang:1.21.6-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent

FROM --platform=$BUILDPLATFORM alpine:3.18.4 AS tools
ARG TARGETPLATFORM
RUN apk add --no-cache curl wget
# Download async-profiler
RUN if [ "$TARGETPLATFORM" = "linux/amd64" ]; then ARCHITECTURE=x64; elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then ARCHITECTURE=arm64; else ARCHITECTURE=x64; fi \
&& wget -O async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz https://github.com/jvm-profiling-tools/async-profiler/releases/download/v${ASYNCP_VERSION}/async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz \
&& tar -xf async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz && mv async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE} async-profiler

FROM --platform=$TARGETPLATFORM amazoncorretto:$JAVA_VERSION_TAG-alpine
RUN mkdir -p /app/async-profiler/build
RUN apk add --no-cache procps
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/profiler.sh /app/async-profiler
COPY --from=tools /async-profiler/build /app/async-profiler/build
RUN chmod +x /app/async-profiler/profiler.sh
RUN mkdir -p /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/jfr-profile.jfc /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]