FROM golang:1.19.2-bullseye as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM alpine as tools
RUN apk add build-base linux-headers curl
# Download async-profiler
RUN curl -o async-profiler-2.8-linux-musl-x64.tar.gz -L \
    https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8/async-profiler-2.8-linux-musl-x64.tar.gz
RUN tar -xvf async-profiler-2.8-linux-musl-x64.tar.gz && mv async-profiler-2.8-linux-musl-x64 async-profiler
# Download jdk-17
RUN curl -o jdk-17.0.3+7.tar.gz -L \
    https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.3+7/OpenJDK17U-jdk_x64_alpine-linux_hotspot_17.0.3_7.tar.gz
RUN tar -xvf jdk-17.0.3+7.tar.gz && mv jdk-17.0.3+7 jdk-17

FROM alpine
RUN mkdir -p /app/async-profiler/build
RUN apk add procps
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/profiler.sh /app/async-profiler
COPY --from=tools /async-profiler/build /app/async-profiler/build
RUN chmod +x /app/async-profiler
COPY --from=tools /jdk-17 /opt/jdk-17

CMD [ "/app/agent" ]