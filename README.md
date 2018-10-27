# Ruthie
Reliable (with [Ami](https://github.com/kak-tus/ami) and [Redis Cluster Streams](https://redis.io/topics/streams-intro)) Clickhouse writer.

There is another ClickHouser writer [Corrie](https://github.com/kak-tus/corrie) based on RabbitMQ. But it is deprecated - performance of RabbitMQ is too slow. We need extra large count of RabbitMQ nodes (shards) to achive perfomance of Redis Cluster.
