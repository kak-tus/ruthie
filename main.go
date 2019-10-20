package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/kak-tus/healthcheck"
	"github.com/kak-tus/ruthie/config"
	"github.com/kak-tus/ruthie/reader"
	"github.com/kak-tus/ruthie/writer"
	"go.uber.org/zap"
)

var rdr *reader.Reader
var wrt *writer.Writer

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	log := logger.Sugar()

	cnf, err := config.NewConfig()
	if err != nil {
		log.Panic(err)
	}

	hlth := healthcheck.NewHandler()
	hlth.Add("/healthcheck", func() (healthcheck.State, string) {
		return healthcheck.StatePassing, "ok"
	})

	rdr, err = reader.NewReader(&cnf.Reader, log)
	if err != nil {
		log.Panic(err)
	}

	wrt, err = writer.NewWriter(&cnf.Writer, log, rdr)
	if err != nil {
		log.Panic(err)
	}

	go wrt.Start()

	hlth.Add("/status", status)

	go http.ListenAndServe(cnf.Healthcheck.Listen, hlth)

	st := make(chan os.Signal, 1)
	signal.Notify(st, os.Interrupt)

	<-st
	log.Info("Stop")

	wrt.Stop()

	_ = log.Sync()

}

func status() (healthcheck.State, string) {
	var wg sync.WaitGroup
	wg.Add(2)

	var rs bool
	go func() {
		rs = rdr.IsAccessible()
		wg.Done()
	}()

	var ws bool
	go func() {
		ws = wrt.IsAccessible()
		wg.Done()
	}()

	wg.Wait()

	if rs && ws {
		return healthcheck.StatePassing, "ok"
	}

	return healthcheck.StateWarning, "nok"
}
