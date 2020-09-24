FROM golang:1.14-alpine as builder
MAINTAINER lhy1024
ENV GO111MODULE=on
WORKDIR /src
COPY . .
RUN go build -o bin/bench *.go


FROM golang:1.14-alpine as pdbuilder

RUN apk add --no-cache \
    make \
    git \
    bash \
    curl \
    gcc \
    g++
   
RUN mkdir -p /go/src/github.com/tikv/pd
WORKDIR /go/src/github.com/tikv/pd

RUN git clone https://github.com/tikv/pd.git .
RUN make simulator


FROM alpine:3.5

WORKDIR /artifacts
RUN mkdir conf
RUN mkdir -p /scripts/simulator

COPY --from=builder /src/bin/* /bin/
COPY --from=builder /src/scripts/simulator/* /scripts/simulator/
COPY --from=pdbuilder /go/src/github.com/tikv/pd/bin/pd-simulator /bin/
COPY --from=pdbuilder /go/src/github.com/tikv/pd/conf/simconfig.toml conf/simconfig.toml

RUN chmod +x /scripts/simulator/*

CMD ["/bin/bench"]
