FROM golang:1.14-alpine as builder
MAINTAINER lhy1024
ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN go build -o bin/bench *.go

FROM alpine:3.5
COPY --from=0 /src/bin/* /bin/
CMD ["/bin/bench"]