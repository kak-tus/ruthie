package reader

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/kak-tus/ami"
	"go.uber.org/zap"
)

func NewReader(cnf *Config, log *zap.SugaredLogger) (*Reader, error) {
	addrs := strings.Split(cnf.Redis.Addrs, ",")

	rdr := &Reader{
		cnf: cnf,
		log: log,
	}

	cn, err := ami.NewConsumer(
		ami.ConsumerOptions{
			Block:             time.Second,
			Consumer:          cnf.Consumer,
			ErrorNotifier:     rdr,
			Name:              cnf.QueueName,
			PendingBufferSize: cnf.PendingBufferSize,
			PipeBufferSize:    cnf.PipeBufferSize,
			PipePeriod:        time.Microsecond * 1000,
			PrefetchCount:     cnf.PrefetchCount,
			ShardsCount:       cnf.ShardsCount,
		},
		&redis.ClusterOptions{
			Addrs:        addrs,
			ReadTimeout:  time.Second * 60,
			WriteTimeout: time.Second * 60,
		},
	)
	if err != nil {
		return nil, err
	}

	rdr.cn = cn

	return rdr, nil
}

func (r *Reader) Start() chan ami.Message {
	r.log.Info("Start reader")
	ch := r.cn.Start()
	r.log.Info("Started reader")
	return ch
}

func (r *Reader) Stop() {
	r.log.Info("Stop reader")
	r.cn.Stop()
	r.cn.Close()
	r.log.Info("Stopped reader")
}

// IsAccessible checks Redis status
func (r *Reader) IsAccessible() bool {
	// TODO ping
	return true
}

// Ack message
func (r *Reader) Ack(m ami.Message) {
	r.cn.Ack(m)
}

func (r *Reader) AmiError(err error) {
	r.log.Error(err)
}
