package writer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"git.aqq.me/go/retrier"
	"github.com/iph0/conf"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/ruthie/message"
	"github.com/kak-tus/ruthie/reader"
	"github.com/kshvakov/clickhouse"
)

var wrt *Writer

func init() {
	event.Init.AddHandler(
		func() error {
			cnfMap := appconf.GetConfig()["writer"]

			var cnf writerConfig
			err := conf.Decode(cnfMap, &cnf)
			if err != nil {
				return err
			}

			db, err := sql.Open("clickhouse", cnf.ClickhouseURI)
			if err != nil {
				return err
			}

			err = db.Ping()
			if err != nil {
				exception, ok := err.(*clickhouse.Exception)
				if ok {
					return fmt.Errorf("[%d] %s \n%s", exception.Code, exception.Message, exception.StackTrace)
				}

				return err
			}

			wrt = &Writer{
				log:        applog.GetLogger().Sugar(),
				cnf:        cnf,
				db:         db,
				dec:        jsoniter.Config{UseNumber: true}.Froze(),
				m:          &sync.Mutex{},
				rdr:        reader.GetReader(),
				toSendCnts: make(map[string]int),
				toSendVals: make(map[string][]*toSend),
				retr:       retrier.New(retrier.Config{RetryPolicy: []time.Duration{time.Second * 5}}),
			}

			wrt.log.Info("Started writer")

			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			wrt.log.Info("Stop writer")

			wrt.retr.Stop()

			wrt.rdr.Stop()

			wrt.m.Lock()
			wrt.db.Close()

			wrt.log.Info("Stopped writer")

			return nil
		},
	)
}

// GetWriter return instance
func GetWriter() *Writer {
	return wrt
}

// Start writer
func (w Writer) Start() {
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
			w.rdr.Ack(msg)
			continue
		}

		if len(parsed.Query) == 0 {
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

		if time.Now().Sub(start) >= w.cnf.Period {
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

		diffSend := time.Now().Sub(started)
		started = time.Now()

		for _, v := range w.toSendVals[query][0:w.toSendCnts[query]] {
			if v.failed {
				w.log.Error("Failed: ", v.msgAmi.Body)
			}

			w.rdr.Ack(v.msgAmi)
		}

		diffAck := time.Now().Sub(started)
		w.log.Infof("Sended %d values in %fsec, acked in %fsec for %q",
			w.toSendCnts[query], diffSend.Seconds(), diffAck.Seconds(), query)

		w.toSendCnts[query] = 0
	}
}

func (w *Writer) send(query string, vals []*toSend) error {
	err := w.retr.Do(func() *retrier.Error {
		tx, err := w.db.Begin()
		if err != nil {
			w.log.Error("Start transaction failed: ", err)
			return retrier.NewError(err, false)
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
			return retrier.NewError(err, false)
		}

		return nil
	})
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
