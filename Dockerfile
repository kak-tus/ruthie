FROM golang:1.13.2-alpine3.10 AS build

WORKDIR /go/ruthie

COPY config ./config
COPY go.mod .
COPY go.sum .
COPY main.go .
COPY message ./message
COPY reader ./reader
COPY writer ./writer

RUN go build -o /go/bin/ruthie

FROM alpine:3.10

COPY --from=build /go/bin/ruthie /usr/local/bin/ruthie
COPY etc /etc/

RUN \
  apk add --no-cache \
    tzdata \
  && adduser -DH user

USER user

ENV \
  RUTHIE_CLICKHOUSE_ADDR= \
  RUTHIE_CLICKHOUSE_ALTADDRS= \
  RUTHIE_CLICKHOUSE_PASSWORD= \
  RUTHIE_CLICKHOUSE_USERNAME= \
  \
  RUTHIE_CONSUMER_NAME= \
  RUTHIE_REDIS_ADDR= \
  \
  RUTHIE_BATCH=10000 \
  RUTHIE_PENDING_BUFFER_SIZE=1000000 \
  RUTHIE_PERIOD=60000000000 \
  RUTHIE_PIPE_BUFFER_SIZE=50000 \
  RUTHIE_PREFETCH_COUNT=30000 \
  RUTHIE_QUEUE_NAME_FAILED=ruthie_failed \
  RUTHIE_QUEUE_NAME=ruthie \
  RUTHIE_SHARDS_COUNT=10

CMD ["/usr/local/bin/ruthie"]
