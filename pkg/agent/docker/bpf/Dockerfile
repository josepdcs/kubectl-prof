ARG KERNEL_VERSION=5.15.15

FROM linuxkit/kernel:$KERNEL_VERSION AS ksrc

FROM golang:1.17.7-buster as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-profile
ADD . /go/src/github.com/josepdcs/kubectl-profile
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM alpine as builder
COPY --from=ksrc /kernel-dev.tar /
RUN tar xf /kernel-dev.tar -C /
RUN mv /usr/src/*/ /usr/src/kernel-source/
RUN apk add git
RUN git clone https://github.com/brendangregg/FlameGraph

FROM alpine
COPY --from=builder /usr/src /usr/src
RUN apk add bcc-tools perl
RUN ln -s $(which python3) /usr/bin/python
RUN mkdir -p /app/FlameGraph
COPY --from=builder /FlameGraph /app/FlameGraph
COPY --from=agentbuild /go/bin/agent /app
COPY --from=agentbuild /go/src/github.com/josepdcs/kubectl-profile/tools/bpf/profile.py /app/bcc-profiler/profile
RUN chmod +x /app/bcc-profiler/profile

CMD [ "/app/agent" ]