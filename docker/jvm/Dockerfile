ARG JAVA_VERSION_TAG=21
FROM --platform=$BUILDPLATFORM golang:1.22.5-bullseye AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent


FROM --platform=$BUILDPLATFORM eclipse-temurin:$JAVA_VERSION_TAG AS tools
ARG TARGETPLATFORM
ARG ASYNCP_VERSION=2.9
RUN apt-get update && apt-get install -y curl tar wget
# Echo variables to debug and ensure they hold the expected values
RUN echo "ASYNCP_VERSION=${ASYNCP_VERSION}, TARGETPLATFORM=${TARGETPLATFORM}"

RUN $JAVA_HOME/bin/jlink \
         --module-path jmods \
         --add-modules jdk.jcmd \
         --strip-debug \
         --no-man-pages \
         --no-header-files \
         --compress=2 \
         --output /jcmd

# Determine the architecture based on the platform
RUN if [ "$TARGETPLATFORM" = "linux/amd64" ]; then ARCHITECTURE=x64; elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then ARCHITECTURE=arm64; else ARCHITECTURE=x64; fi \
&& wget -O async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz https://github.com/async-profiler/async-profiler/releases/download/v${ASYNCP_VERSION}/async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz \
&& tar -xf async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE}.tar.gz && mv async-profiler-${ASYNCP_VERSION}-linux-${ARCHITECTURE} async-profiler

FROM --platform=$TARGETPLATFORM bitnami/minideb:bullseye
RUN apt-get update && apt-get install -y procps strace sudo gzip
RUN mkdir -p /app/async-profiler/build
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/profiler.sh /app/async-profiler
COPY --from=tools /async-profiler/build /app/async-profiler/build
COPY --from=tools /jcmd /opt/jdk
RUN chmod +x /app/async-profiler/profiler.sh
RUN mkdir -p /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/jfr-profile.jfc /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]
