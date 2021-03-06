package writer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/ami"
	"github.com/kak-tus/ruthie/message"
	"github.com/kak-tus/ruthie/reader"
	"github.com/kshvakov/clickhouse"
	"github.com/ssgreg/repeat"
	"go.uber.org/zap"
)

func NewWriter(cnf *Config, log *zap.SugaredLogger, rdr *reader.Reader) (*Writer, error) {
	db, err := sql.Open("clickhouse", cnf.ClickhouseURI)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		exception, ok := err.(*clickhouse.Exception)
		if ok {
			return nil, fmt.Errorf("[%d] %s \n%s", exception.Code, exception.Message, exception.StackTrace)
		}

		return nil, err
	}

	addrs := strings.Split(cnf.Redis.Addrs, ",")

	wrt := &Writer{
		log:        log,
		cnf:        cnf,
		db:         db,
		dec:        jsoniter.Config{UseNumber: true}.Froze(),
		m:          &sync.Mutex{},
		toSendCnts: make(map[string]int),
		toSendVals: make(map[string][]*toSend),
		rdr:        rdr,
	}

	prod, err := ami.NewProducer(
		ami.ProducerOptions{
			ErrorNotifier:     wrt,
			Name:              cnf.QueueNameFailed,
			PendingBufferSize: 10000000,
			PipeBufferSize:    50000,
			PipePeriod:        time.Microsecond * 1000,
			ShardsCount:       10,
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

	wrt.prod = prod

	return wrt, nil
}

// Start writer
func (w *Writer) Start() {
	w.log.Info("Start writer")

	w.m.Lock()
	defer w.m.Unlock()

	c := w.rdr.Start()

	start := time.Now()

	for {
		msg, more := <-c
		if !more {
			w.sendAll()
			break
		}

		var parsed message.Message
		err := w.dec.UnmarshalFromString(msg.Body, &parsed)
		if err != nil {
			w.log.Error("Decode failed: ", err)
			w.prod.Send(msg.Body)
			w.rdr.Ack(msg)
			continue
		}

		if len(parsed.Query) == 0 {
			w.prod.Send(msg.Body)
			w.rdr.Ack(msg)
			continue
		}

		if w.toSendVals[parsed.Query] == nil {
			w.toSendVals[parsed.Query] = make([]*toSend, w.cnf.Batch)
			w.toSendCnts[parsed.Query] = 0
		}

		w.toSendVals[parsed.Query][w.toSendCnts[parsed.Query]] = &toSend{
			msgParsed: parsed,
			msgAmi:    msg,
			failed:    false,
		}

		w.toSendCnts[parsed.Query]++

		if w.toSendCnts[parsed.Query] >= w.cnf.Batch {
			w.sendOne(parsed.Query)
		}

		if time.Since(start) >= w.cnf.Period {
			w.sendAll()
			start = time.Now()
		}
	}
}

// IsAccessible checks Clickhouse status
func (w Writer) IsAccessible() bool {
	for i := 0; i < 10; i++ {
		err := w.db.Ping()
		if err == nil {
			return true
		}

		w.log.Error("Ping failed: ", err)
		time.Sleep(time.Second)
	}

	return false
}

func (w *Writer) sendAll() {
	for query := range w.toSendVals {
		w.sendOne(query)
	}
}

func (w *Writer) sendOne(query string) {
	if w.toSendCnts[query] > 0 {
		started := time.Now()
		err := w.send(query, w.toSendVals[query][0:w.toSendCnts[query]])
		if err != nil {
			// If we got error here - this means, that stop of the service
			// is requested. Because retrier do retries infinitely, while stopped.
			w.log.Error(err)
			w.toSendCnts[query] = 0
			return
		}

		diffSend := time.Since(started)
		started = time.Now()

		for _, v := range w.toSendVals[query][0:w.toSendCnts[query]] {
			if v.failed {
				w.log.Error("Failed: ", v.msgAmi.Body)
				w.prod.Send(v.msgAmi.Body)
			}

			w.rdr.Ack(v.msgAmi)
		}

		diffAck := time.Since(started)
		w.log.Infof("Sended %d values in %fsec, acked in %fsec for %q",
			w.toSendCnts[query], diffSend.Seconds(), diffAck.Seconds(), query)

		w.toSendCnts[query] = 0
	}
}

func (w *Writer) send(query string, vals []*toSend) error {
	err := repeat.Repeat(
		repeat.Fn(func() error {
			tx, err := w.db.Begin()
			if err != nil {
				w.log.Error("Start transaction failed: ", err)
				return repeat.HintTemporary(err)
			}

			stmt, err := tx.Prepare(query)
			if err != nil {
				w.log.Error("Prepare query failed: ", err)

				err = tx.Rollback()
				if err != nil {
					w.log.Error("Rollback failed: ", err)
				}

				for _, v := range vals {
					v.failed = true
				}

				return nil
			}

			// There is no need to commit if no one succeeded exec
			succeded := 0

			for _, v := range vals {
				if v.failed {
					continue
				}

				data := w.makeCHArray(v.msgParsed.Data)
				_, err := stmt.Exec(data...)

				if err != nil {
					w.log.Error("Exec failed: ", err)
					v.failed = true
					continue
				}

				succeded++
			}

			if succeded == 0 {
				err := tx.Rollback()
				if err != nil {
					w.log.Error("Rollback failed: ", err)
				}
				return nil
			}

			err = tx.Commit()
			if err != nil {
				w.log.Error("Commit failed: ", err)
				return repeat.HintTemporary(err)
			}

			return nil
		}),
		repeat.StopOnSuccess(),
		repeat.WithDelay(repeat.FullJitterBackoff(500*time.Millisecond).Set()),
	)

	if err != nil {
		// If we got error here - this means, that stop of the service
		// is requested. Because retrier do retries infinitely, while stopped.
		return err
	}

	return nil
}

func (w Writer) makeCHArray(vals []interface{}) []interface{} {
	data := make([]interface{}, len(vals))

	for i, v := range vals {
		num, ok := v.(json.Number)

		if !ok {
			data[i] = v
			continue
		}

		convI, err := num.Int64()
		if err == nil {
			data[i] = convI
			continue
		}

		convF, err := num.Float64()
		if err == nil {
			data[i] = convF
			continue
		}

		data[i] = v
	}

	return data
}

func (w *Writer) AmiError(err error) {
	w.log.Error(err)
}

func (w *Writer) Stop() {
	w.log.Info("Stop writer")

	w.rdr.Stop()

	w.prod.Close()

	w.m.Lock()
	w.db.Close()

	w.log.Info("Stopped writer")
}
