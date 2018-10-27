package message

import jsoniter "github.com/json-iterator/go"

var decoder = jsoniter.Config{UseNumber: true}.Froze()

// Message structure
type Message struct {
	Query string
	Data  []interface{}
}

// Encode message
func (m Message) Encode() (string, error) {
	return decoder.MarshalToString(m)
}
