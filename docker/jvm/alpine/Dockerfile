ARG JAVA_VERSION_TAG=21
ARG ASYNCP_VERSION=2.9
FROM --platform=$BUILDPLATFORM golang:1.22.5-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent

FROM --platform=$BUILDPLATFORM amazoncorretto:11-alpine-jdk AS async-profiler-builder
ARG TARGETPLATFORM
ARG ASYNCP_VERSION=2.9
RUN apk add --no-cache curl tar wget g++ linux-headers musl-dev git make
WORKDIR /src
RUN git clone --branch v${ASYNCP_VERSION} --depth 1 https://github.com/async-profiler/async-profiler
WORKDIR /src/async-profiler
RUN make CXXFLAGS='-O3 -fno-omit-frame-pointer -fvisibility=hidden -D_LARGEFILE64_SOURCE' all native; \
    tree -d /src;

FROM --platform=$BUILDPLATFORM eclipse-temurin:$JAVA_VERSION_TAG-alpine AS tools
ARG TARGETPLATFORM
ARG ASYNCP_VERSION=2.9
RUN apk add --no-cache curl tar wget g++ linux-headers git make
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

FROM --platform=$TARGETPLATFORM alpine:3.18.4
RUN mkdir -p /app/async-profiler/build
RUN apk add --no-cache procps coreutils
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/profiler.sh /app/async-profiler
COPY --from=async-profiler-builder /src/async-profiler/build /app/async-profiler/build
COPY --from=tools /jcmd /opt/jdk
RUN chmod +x /app/async-profiler/profiler.sh
RUN mkdir -p /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/jfr-profile.jfc /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]