package main

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/kak-tus/ami"
	"github.com/kak-tus/ruthie/message"
)

func main() {
	qu, err := ami.NewProducer(
		ami.ProducerOptions{
			ErrorNotifier:     &errorLogger{},
			Name:              "ruthie",
			PendingBufferSize: 10000000,
			PipeBufferSize:    50000,
			PipePeriod:        time.Microsecond * 1000,
			ShardsCount:       10,
		},
		&redis.ClusterOptions{
			Addrs:        []string{"172.17.0.1:7001", "172.17.0.1:7002"},
			ReadTimeout:  time.Second * 60,
			WriteTimeout: time.Second * 60,
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

type errorLogger struct{}

func (l *errorLogger) AmiError(err error) {
	println("Got error from Ami:", err.Error())
}
