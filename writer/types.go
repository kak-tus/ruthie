package writer

import (
	"database/sql"
	"sync"
	"time"

	"git.aqq.me/go/retrier"
	jsoniter "github.com/json-iterator/go"
	"github.com/kak-tus/ami"
	"github.com/kak-tus/ruthie/message"
	"github.com/kak-tus/ruthie/reader"
	"go.uber.org/zap"
)

// Writer hold object
type Writer struct {
	cnf        writerConfig
	db         *sql.DB
	dec        jsoniter.API
	log        *zap.SugaredLogger
	m          *sync.Mutex
	rdr        *reader.Reader
	retr       *retrier.Retrier
	toSendCnts map[string]int
	toSendVals map[string][]*toSend
}

type writerConfig struct {
	Batch         int
	ClickhouseURI string
	Period        time.Duration
}

type toSend struct {
	failed    bool
	msgAmi    ami.Message
	msgParsed message.Message
}
