package bztcp

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"time"
)

// Conn implements a Benzinga TCP connection.
type Conn struct {
	socket net.Conn
	reader *bufio.Reader
}

// Dial connects to a Benzinga TCP server.
func Dial(addr, user, key string) (*Conn, error) {
	return DialTimeout(addr, user, key, AuthTimeout)
}

// DialTLS connects to a Benzinga TCP server using TLS.
func DialTLS(addr, user, key string) (*Conn, error) {
	return DialTimeoutTLS(addr, user, key, AuthTimeout)
}

// DialTimeout connects to Benzinga TCP with a timeout.
func DialTimeout(addr, user, key string, d time.Duration) (*Conn, error) {
	socket, err := net.DialTimeout("tcp", addr, d)

	if err != nil {
		return nil, err
	}

	return NewConn(socket, user, key)
}

// DialTimeoutTLS connects to Benzinga TCP with a timeout using TLS.
func DialTimeoutTLS(addr, user, key string, d time.Duration) (*Conn, error) {
	socket, err := tls.DialWithDialer(&net.Dialer{Timeout: d}, "tcp", addr, nil)

	if err != nil {
		return nil, err
	}

	return NewConn(socket, user, key)
}

// NewConn connects to TCP using an already-configured socket.
func NewConn(socket net.Conn, user, key string) (*Conn, error) {
	conn := Conn{
		socket: socket,
		reader: bufio.NewReader(socket),
	}

	// Attempt to authenticate with credentials.
	err := conn.authenticate(user, key)
	if err != nil {
		socket.Close()
		return nil, err
	}

	// Attempt to enable TCP keep alive.
	if tcpconn, ok := socket.(*net.TCPConn); ok {
		tcpconn.SetKeepAlive(true)
	}

	return &conn, nil
}

// Stream watches the connection, calling the cb function whenever it
// encounters a stream message. This function will exit when the connection
// closes, or when the context is closed. If the context is closed, the
// connection will be closed.
func (c *Conn) Stream(ctx context.Context, cb func(d StreamData)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Ping thread
	go func() {
		t := time.NewTicker(PingDuration)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				c.Send("PING", PingData{time.Now().Format(TimeFormat)})
			case <-ctx.Done():
				c.socket.Close()
				return
			}
		}
	}()

	// Read loop
	for {
		// Read next message.
		msg, err := c.Recv()

		// Discard error if the context was closed while reading.
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err != nil {
			return err
		}

		// Interpret and handle message.
		switch msg.Status {
		case "PONG":
			// do nothing
		case "STREAM":
			data := StreamData{}
			err := json.Unmarshal(msg.Data, &data)

			if err != nil {
				return err
			}

			cb(data)
		}
	}
}

// Recv gets the next message in the stream. This function is low-level and
// not necessary; most users should use Stream instead.
func (c *Conn) Recv() (Message, error) {
	// Read line from TCP socket.
	message := Message{}
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}

	// Deserialize message.
	err = message.Decode(line)
	if err != nil {
		return Message{}, err
	}

	return message, nil
}

// Send sends a message to the server. This function is low-level and not
// usually necessary; most users should use Stream instead.
func (c *Conn) Send(status string, body interface{}) error {
	// Serialize message body.
	msg, err := NewMessage(status, body)
	if err != nil {
		return err
	}

	// Encode to socket.
	_, err = c.socket.Write(msg.Encode())
	if err != nil {
		return err
	}

	return nil
}

// authenticate handles the authentication handshake.
//
// The exchange generally looks like this:
//
//     < READY=BZEOT
//     > AUTH: {"username":"bztest","key":"12345"}=BZEOT
//     < CONNECTED=BZEOT
//
// In error cases, such as an invalid key, you may see
// an error response instead of `CONNECTED`:
//
//   - "INVALID KEY FORMAT": An error occurred decoding the authdata.
//   - "INVALID KEY": The user or key was not valid.
//
// An example of such transmission follows:
//
//     < READY=BZEOT
//     > AUTH: {"username":"bztest",}=BZEOT
//     < INVALID KEY FORMAT=BZEOT
//
func (c *Conn) authenticate(user, key string) error {
	c.socket.SetDeadline(time.Now().Add(AuthTimeout))

	// Read 'READY' message.
	msg, err := c.Recv()
	if err != nil {
		return err
	} else if msg.Status != "READY" {
		return ErrInvalidReady
	}

	// Write 'AUTH' message.
	err = c.Send("AUTH", AuthData{
		Username: user,
		Key:      key,
	})
	if err != nil {
		return err
	}

	// Read 'AUTH' response message.
	msg, err = c.Recv()
	if err != nil {
		return err
	}

	// Clear the deadline.
	c.socket.SetDeadline(time.Time{})

	// Handle message status.
	switch msg.Status {
	case "INVALID KEY FORMAT":
		return ErrInvalidKeyFormat
	case "INVALID KEY":
		return ErrInvalidKey
	case "CONNECTED":
		return nil
	default:
		return ErrInvalidAuthResponse
	}
}
