FROM golang:1.19.6-alpine as agentbuild
WORKDIR /go/src/github.com/josepdcs/kubectl-prof
ADD . /go/src/github.com/josepdcs/kubectl-prof
RUN go get -d -v ./...
RUN cd cmd/agent && go build -o /go/bin/agent

FROM python:3.11.2-slim-bullseye AS pyspybuild
RUN pip3 install py-spy==0.3.11

FROM python:3.11.2-alpine
RUN apk add --update procps strace
RUN mkdir /app
COPY --from=agentbuild /go/bin/agent /app/agent
COPY --from=pyspybuild /usr/local/bin/py-spy /app/py-spy

CMD [ "/app/agent" ]