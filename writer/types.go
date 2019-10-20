package writer

import (
	"database/sql"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/ami"
	"github.com/kak-tus/ruthie/message"
	"github.com/kak-tus/ruthie/reader"
	"go.uber.org/zap"
)

// Writer hold object
type Writer struct {
	cnf        *Config
	db         *sql.DB
	dec        jsoniter.API
	log        *zap.SugaredLogger
	m          *sync.Mutex
	prod       *ami.Producer
	rdr        *reader.Reader
	toSendCnts map[string]int
	toSendVals map[string][]*toSend
}

type Config struct {
	Batch           int
	ClickhouseURI   string
	Period          time.Duration
	QueueNameFailed string
	Redis           redisConfig
}

type toSend struct {
	failed    bool
	msgAmi    ami.Message
	msgParsed message.Message
}

type redisConfig struct {
	Addrs string
}
