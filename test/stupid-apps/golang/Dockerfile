FROM golang:1.21.5-bullseye AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app
COPY ./test/stupid-apps/golang/go.mod /app/go.mod
COPY ./test/stupid-apps/golang/main.go /app/main.go
RUN go get -d -v ./...
RUN go build -o poc

FROM scratch
COPY --from=builder /app /app

CMD [ "/app/poc" ]
