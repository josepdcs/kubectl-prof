ARG JAVA_VERSION_TAG=21
ARG ASYNCP_VERSION=2.9
FROM golang:1.24.5-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
COPY . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/agent

FROM amazoncorretto:11-alpine-jdk AS async-profiler-builder
ARG TARGETPLATFORM
ARG ASYNCP_VERSION=2.9
RUN apk add --no-cache curl g++ git linux-headers make musl-dev tar wget
WORKDIR /src
RUN git clone --branch v${ASYNCP_VERSION} --depth 1 https://github.com/async-profiler/async-profiler
WORKDIR /src/async-profiler
RUN make CXXFLAGS='-O3 -fno-omit-frame-pointer -fvisibility=hidden -D_LARGEFILE64_SOURCE' all native; \
    tree -d /src;

FROM eclipse-temurin:$JAVA_VERSION_TAG-alpine AS tools
ARG TARGETPLATFORM
ARG ASYNCP_VERSION=2.9
RUN apk add --no-cache curl g++ git linux-headers make tar wget  \
    && echo "ASYNCP_VERSION=${ASYNCP_VERSION}, TARGETPLATFORM=${TARGETPLATFORM}" \
    && "$JAVA_HOME"/bin/jlink \
         --module-path jmods \
         --add-modules jdk.jcmd \
         --strip-debug \
         --no-man-pages \
         --no-header-files \
         --compress=2 \
         --output /jcmd

FROM alpine:3.22.1
RUN mkdir -p /app/async-profiler/build  \
    && apk add --no-cache procps coreutils
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/profiler.sh /app/async-profiler
COPY --from=async-profiler-builder /src/async-profiler/build /app/async-profiler/build
COPY --from=tools /jcmd /opt/jdk
RUN chmod +x /app/async-profiler/profiler.sh  \
    && mkdir -p /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/jfr-profile.jfc /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]