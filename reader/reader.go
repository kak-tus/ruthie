package reader

import (
	"strings"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"github.com/go-redis/redis"
	"github.com/iph0/conf"
	"github.com/kak-tus/ami"
)

var rdr *Reader

func init() {
	event.Init.AddHandler(
		func() error {
			cnfMap := appconf.GetConfig()["reader"]

			var cnf readerConfig
			err := conf.Decode(cnfMap, &cnf)
			if err != nil {
				return err
			}

			addrs := strings.Split(cnf.Redis.Addrs, ",")

			cn, err := ami.NewConsumer(
				ami.ConsumerOptions{
					Name:              cnf.QueueName,
					Consumer:          cnf.Consumer,
					ShardsCount:       cnf.ShardsCount,
					PrefetchCount:     cnf.PrefetchCount,
					Block:             time.Second,
					PendingBufferSize: cnf.PendingBufferSize,
					PipeBufferSize:    cnf.PipeBufferSize,
					PipePeriod:        time.Microsecond * 1000,
				},
				&redis.ClusterOptions{
					Addrs: addrs,
				},
			)
			if err != nil {
				return err
			}

			rdr = &Reader{
				log: applog.GetLogger().Sugar(),
				cnf: cnf,
				cn:  cn,
			}

			rdr.log.Info("Started reader")

			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			rdr.log.Info("Stop reader")
			rdr.cn.Close()
			rdr.log.Info("Stopped reader")
			return nil
		},
	)
}

// GetReader return instance
func GetReader() *Reader {
	return rdr
}

// Start reader
func (r *Reader) Start() chan ami.Message {
	return r.cn.Start()
}

// IsAccessible checks Redis status
func (r *Reader) IsAccessible() bool {
	// TODO ping
	return true
}

// Stop reader
func (r *Reader) Stop() {
	r.log.Info("Stop consumer")
	r.cn.Stop()
	r.log.Info("Stopped consumer")
}

// Ack message
func (r *Reader) Ack(m ami.Message) {
	r.cn.Ack(m)
}
