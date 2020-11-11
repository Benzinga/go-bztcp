package bztcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const (
	// TimeFormat is how timestamps are expected to be formatted throughout most of
	// the protocol.
	TimeFormat = "Mon Jan _2 2006 15:04:05 GMT-0700 (MST)"

	// EOL is the end-of-line magic for the BZ TCP protocol.
	EOL = "=BZEOT\r\n"

	// AuthTimeout is the amount of time spent waiting for authentication.
	AuthTimeout = 10 * time.Second

	// PingDuration is the amount of time between sending pings.
	PingDuration = 20 * time.Second
)

var (
	eol = []byte(EOL)
	cdt = []byte(": ")
)

// Message is a raw message from the BZ TCP service.
type Message struct {
	Status string
	Data   json.RawMessage
}

// AuthData is the message data sent in the AUTH message.
type AuthData struct {
	Username string `json:"username"`
	Key      string `json:"key"`
}

// PingData is the message data sent in the PING message.
type PingData struct {
	PingTime string `json:"pingTime"`
}

// PongData is the message data sent in the PONG message.
type PongData struct {
	ServerTime string `json:"serverTime,omitempty"`
	PingTime   string `json:"pingTime"`
}

// Author represents an author
type Author struct {
	Name string `json:"name"`
}

// TickerData contains a symbol associated with a STREAM message.
type TickerData struct {
	Name      string `json:"name"`
	Extended  bool   `json:"-"`
	Primary   bool   `json:"primary"`
	Sentiment int    `json:"sentiment"`
}

// Ticker is a special type that handles extended tickers.
type Ticker TickerData

// StreamData is the message data sent in the STREAM message.
type StreamData struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Body        string      `json:"body"`
	Authors     []Author    `json:"authors,omitempty"`
	PublishedAt string      `json:"published"`
	UpdatedAt   string      `json:"updated"`
	Channels    []string    `json:"channels"`
	Tickers     []Ticker    `json:"tickers"`
	Status      string      `json:"status"`
	Link        interface{} `json:"link"`
}

// UnmarshalJSON implements json.Unmarshaler
func (t *Ticker) UnmarshalJSON(b []byte) error {
	switch b[0] {
	case '{':
		return json.Unmarshal(b, (*TickerData)(t))
	case '"':
		return json.Unmarshal(b, &t.Name)
	default:
		return fmt.Errorf("unexpected character '%c' in ticker", b[0])
	}
}

// NewMessage creates a new messsage with the provided data.
func NewMessage(status string, body interface{}) (Message, error) {
	msg := Message{Status: status}

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return Message{}, err
		}

		msg.Data = data
	}

	return msg, nil
}

// Decode parses a message from bytes.
func (m *Message) Decode(line []byte) error {
	colindex := bytes.IndexAny(line, ":=")
	if colindex == -1 {
		return fmt.Errorf("invalid line: %q", line)
	}

	m.Status = string(line[0:colindex])
	equalidx := bytes.LastIndexByte(line, '=')

	if colindex < equalidx {
		m.Data = bytes.TrimSpace(line[colindex+1 : equalidx])
	} else {
		m.Data = nil
	}

	return nil
}

// Encode translates the message into bytes.
func (m *Message) Encode() []byte {
	status, data := []byte(m.Status), []byte(m.Data)
	buffer := bytes.Buffer{}
	buffer.Grow(len(status) + len(data) + 10)
	buffer.Write(status)
	if m.Data != nil {
		buffer.Write(cdt)
		buffer.Write([]byte(data))
	}
	buffer.Write(eol)
	return buffer.Bytes()
}
