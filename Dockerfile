FROM golang:1.10.3-alpine3.8 AS build

WORKDIR /go/src/github.com/kak-tus/ruthie

COPY message ./message
COPY reader ./reader
COPY vendor ./vendor
COPY writer ./writer
COPY main.go .

RUN go install

FROM alpine:3.8

COPY --from=build /go/bin/ruthie /usr/local/bin/ruthie
COPY etc /etc/

RUN adduser -DH user

USER user

ENV \
  RUTHIE_CLICKHOUSE_ADDR= \
  RUTHIE_CLICKHOUSE_ALTADDRS= \
  \
  RUTHIE_REDIS_ADDR= \
  RUTHIE_CONSUMER_NAME= \
  \
  RUTHIE_BATCH=10000 \
  RUTHIE_PERIOD=60000000000 \
  RUTHIE_SHARDS_COUNT=10 \
  RUTHIE_PREFETCH_COUNT=30000 \
  RUTHIE_PENDING_BUFFER_SIZE=1000000 \
  RUTHIE_PIPE_BUFFER_SIZE=50000

CMD ["/usr/local/bin/ruthie"]
