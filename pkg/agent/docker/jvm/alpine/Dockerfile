FROM golang:1.17.7-buster as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-profile
ADD . /go/src/github.com/josepdcs/kubectl-profile
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent


FROM adoptopenjdk/openjdk11:alpine as asyncprofiler
RUN apk add build-base linux-headers curl
RUN curl -o async-profiler-2.7-linux-musl-x64.tar.gz -L \
    https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.7/async-profiler-2.7-linux-musl-x64.tar.gz
RUN tar -xvf async-profiler-2.7-linux-musl-x64.tar.gz && mv async-profiler-2.7-linux-musl-x64 async-profiler

FROM alpine
RUN mkdir -p /app/async-profiler/build
COPY --from=agentbuild /go/bin/agent /app
COPY --from=asyncprofiler /async-profiler/build /app/async-profiler/build
COPY --from=asyncprofiler /async-profiler/profiler.sh /app/async-profiler
CMD [ "/app/agent" ]