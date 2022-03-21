FROM golang:1.17.7-buster as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM alpine as asyncprofiler
RUN apk add build-base linux-headers curl
RUN curl -o async-profiler-2.7-linux-musl-x64.tar.gz -L \
    https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.7/async-profiler-2.7-linux-musl-x64.tar.gz
RUN tar -xvf async-profiler-2.7-linux-musl-x64.tar.gz && mv async-profiler-2.7-linux-musl-x64 async-profiler

FROM alpine
RUN mkdir -p /app/async-profiler/build
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-prof/tools/jvm/profiler.sh /app/async-profiler
COPY --from=asyncprofiler /async-profiler/build /app/async-profiler/build
RUN chmod +x /app/async-profiler

CMD [ "/app/agent" ]