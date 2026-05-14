package protocol

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// MaxMessageBytes limits one newline-delimited protocol message.
const MaxMessageBytes = 1024 * 1024

// Codec reads and writes newline-delimited JSON protocol messages.
type Codec struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

// NewCodec wraps a bidirectional stream with protocol message encoding.
func NewCodec(conn io.ReadWriter) *Codec {
	return &Codec{
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

// WriteMessage serializes one message and flushes it to the stream.
func (c *Codec) WriteMessage(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if len(data) > MaxMessageBytes {
		return fmt.Errorf("message too large: %d bytes", len(data))
	}

	data = append(data, '\n')

	if _, err := c.writer.Write(data); err != nil {
		return err
	}

	return c.writer.Flush()
}

// ReadMessage reads one newline-delimited message from the stream.
func (c *Codec) ReadMessage() (Message, error) {
	var line []byte

	for {
		fragment, err := c.reader.ReadSlice('\n')
		if len(line)+len(fragment) > MaxMessageBytes {
			return Message{}, fmt.Errorf("message too large: %d bytes", len(line)+len(fragment))
		}
		line = append(line, fragment...)

		if err == nil {
			break
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}
