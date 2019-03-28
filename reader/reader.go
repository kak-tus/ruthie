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

			rdr = &Reader{
				cnf: cnf,
				log: applog.GetLogger().Sugar(),
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
				return err
			}

			rdr.cn = cn

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

func (r *Reader) AmiError(err error) {
	r.log.Error(err)
}
