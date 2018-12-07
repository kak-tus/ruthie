# Ruthie

Reliable (with [Ami](https://github.com/kak-tus/ami) and [Redis Cluster Streams](https://redis.io/topics/streams-intro)) ClickHouse writer.

There is another ClickHouser writer [Corrie](https://github.com/kak-tus/corrie) based on RabbitMQ. But it is deprecated - performance of RabbitMQ is too slow. We need extra large count of RabbitMQ nodes (shards) to achive perfomance of Redis Cluster.

## Configuration

### RUTHIE_CLICKHOUSE_ADDR

Primary ClickHouse address in host:port form

```
RUTHIE_CLICKHOUSE_ADDR=clickhouse1.example.com:9000
```

### RUTHIE_CLICKHOUSE_ALTADDRS

Comma separated list of alternative ClickHouse addresses to loadbalancing. Can be empty

```
RUTHIE_CLICKHOUSE_ALTADDRS=clickhouse2.example.com:9000
```

### RUTHIE_REDIS_ADDR

Comma-separated adresses of Redis Cluster in host:port format

```
RUTHIE_REDIS_ADDR=redis-cluster.example.com:7000
```

### RUTHIE_CONSUMER_NAME

Unique consumer name per queue in Redis Cluster.

```
RUTHIE_CONSUMER_NAME=alice
```

### RUTHIE_BATCH

Set batch size of ClickHouse writes.

```
RUTHIE_BATCH=10000
```

### RUTHIE_PERIOD

Set maximum period in microseconds to write to ClickHouse if batch is not fully completed.

```
RUTHIE_PERIOD=60000000000
```

### RUTHIE_SHARDS_COUNT

Ami queues spreads along cluster by default Redis Cluster ability - shards. Every queue has setuped number of streams with same name, but with different shard number. So different streams are placed at different Redis Cluster nodes.

So bigger value get better spreading of queue along cluster. But huge value is not better idea - it got bigger memory usage. Normal value for cluster with 5 masters and 5 slaves - from 5 to 10.

May be later will be added auto-sharding option to place queue on each Redis Cluster node.

Shards count must have identical values in all producers and consumers of this queue.

```
RUTHIE_SHARDS_COUNT=10
```

### RUTHIE_PREFETCH_COUNT

Maximum amount of messages that can be read from queue at same time.

Bigger value get better perfomance, but RUTHIE_BATCH must be bigger too if you setup greater value.

```
RUTHIE_PREFETCH_COUNT=30000
```

### RUTHIE_PENDING_BUFFER_SIZE

Buffer size to acknowledging messages in Redis. Bigger value get bigger memory usase and a little bit better perfomance.

```
RUTHIE_PENDING_BUFFER_SIZE=1000000
```

### RUTHIE_PIPE_BUFFER_SIZE

Request to Redis sended in pipe mode with RUTHIE_PIPE_BUFFER_SIZE numbers of requests in one batch. Bigger value get better perfomance.

```
RUTHIE_PIPE_BUFFER_SIZE=50000
```

## Run

```
docker run --rm -it kaktuss/ruthie
```

## Write data

To write data use [message](https://godoc.org/github.com/kak-tus/ruthie/message) package and [Ami client](https://github.com/kak-tus/ami).
