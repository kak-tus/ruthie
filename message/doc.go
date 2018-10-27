/*
Package message - message object for Ruthie - reliable (with [Ami](https://github.com/kak-tus/ami) and [Redis Cluster Streams](https://redis.io/topics/streams-intro)) Clickhouse writer.

Usage example

	package main

	import (
		"time"

		"github.com/go-redis/redis"
		"github.com/kak-tus/ami"
		"github.com/kak-tus/ruthie/message"
	)

	func main() {
		qu, err := ami.NewQu(
			ami.Options{
				Name:              "ruthie",
				Consumer:          "alice",
				ShardsCount:       10,
				PrefetchCount:     100,
				Block:             time.Second,
				PendingBufferSize: 10000000,
				PipeBufferSize:    50000,
				PipePeriod:        time.Microsecond * 1000,
			},
			&redis.ClusterOptions{
				Addrs: []string{"172.17.0.1:7001", "172.17.0.1:7002"},
			},
		)
		if err != nil {
			panic(err)
		}

		body, err := message.Message{
			Query: "INSERT INTO default.test (some_field) VALUES (?);",
			Data:  []interface{}{1},
		}.Encode()

		if err != nil {
			panic(err)
		}

		qu.Send(body)
		qu.Close()
	}
}

*/
package message
