package reader

import (
	"github.com/kak-tus/ami"
	"go.uber.org/zap"
)

// Reader hold object
type Reader struct {
	log *zap.SugaredLogger
	cnf readerConfig
	qu  *ami.Qu
	c   chan ami.Message
}

type readerConfig struct {
	Redis             redisConfig
	Consumer          string
	ShardsCount       int8
	PrefetchCount     int64
	PendingBufferSize int64
	PipeBufferSize    int64
}

type redisConfig struct {
	Addrs string
}
