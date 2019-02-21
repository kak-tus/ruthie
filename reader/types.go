package reader

import (
	"github.com/kak-tus/ami"
	"go.uber.org/zap"
)

// Reader hold object
type Reader struct {
	c   chan ami.Message
	cn  *ami.Consumer
	cnf readerConfig
	log *zap.SugaredLogger
}

type readerConfig struct {
	Consumer          string
	PendingBufferSize int64
	PipeBufferSize    int64
	PrefetchCount     int64
	QueueName         string
	Redis             redisConfig
	ShardsCount       int8
}

type redisConfig struct {
	Addrs string
}
