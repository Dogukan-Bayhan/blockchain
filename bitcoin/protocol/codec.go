package protocol

import (
	"bufio"
	"encoding/json"
	"io"
)


type Codec struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewCodec(conn io.ReadWriter) *Codec {
	return &Codec{
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (c *Codec) WriteMessage(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if _, err := c.writer.Write(data); err != nil {
		return err
	}

	return c.writer.Flush()
}

func (c *Codec) ReadMessage() (Message, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}

