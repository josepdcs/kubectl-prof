FROM golang:1.21.6-alpine AS agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
WORKDIR /go/src/github.com/josepdcs/kubectl-prof/cmd/agent
RUN go build -o /go/bin/agent

FROM alpine:3.18.4 AS tools
RUN apk add --no-cache curl
# Download async-profiler
RUN curl -o async-profiler-2.8-linux-musl-x64.tar.gz -L \
    https://github.com/async-profiler/async-profiler/releases/download/v2.9/async-profiler-2.9-linux-musl-x64.tar.gz
RUN tar -xvf async-profiler-2.8-linux-musl-x64.tar.gz && mv async-profiler-2.9-linux-musl-x64 async-profiler
# Download jdk
RUN curl -o jdk-21.0.1+12.tar.gz -L \
    https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.1+12/OpenJDK21U-jdk_x64_alpine-linux_hotspot_21.0.1_12.tar.gz
RUN tar -xf jdk-21.0.1+12.tar.gz && mv jdk-21.0.1+12 jdk

FROM alpine:3.18.4
RUN mkdir -p /app/async-profiler/build
RUN apk add --no-cache procps
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/profiler.sh /app/async-profiler
COPY --from=tools /async-profiler/build /app/async-profiler/build
RUN chmod +x /app/async-profiler/profiler.sh
RUN mkdir -p /app/jfr/settings
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/jvm/jfr-profile.jfc /app/jfr/settings
COPY --from=tools /jdk /opt/jdk
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/contrib/get-ps-command.sh /app/get-ps-command.sh
RUN chmod +x /app/get-ps-command.sh

CMD [ "/app/agent" ]