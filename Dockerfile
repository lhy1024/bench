FROM golang:1.14-alpine as builder
MAINTAINER lhy1024
RUN mkdir -p /go/src/github.com/lhy1024/bench
WORKDIR /go/src/github.com/lhy1024/bench
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on go mod download
COPY . .
RUN GO111MODULE=on go build -o bench

FROM alpine:3.5
COPY --from=builder /go/src/github.com/lhy1024/bench /bench
ENTRYPOINT ["/bench"]