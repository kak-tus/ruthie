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
	log        *zap.SugaredLogger
	cnf        writerConfig
	db         *sql.DB
	dec        jsoniter.API
	m          *sync.Mutex
	rdr        *reader.Reader
	toSendVals map[string][]*toSend
	toSendCnts map[string]int
	retr       *retrier.Retrier
}

type writerConfig struct {
	ClickhouseURI string
	Batch         int
	Period        time.Duration
}

type toSend struct {
	msgParsed message.Message
	msgAmi    ami.Message
	failed    bool
}
