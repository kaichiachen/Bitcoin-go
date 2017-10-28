package bitcoin

import (
	"bytes"
	"errors"
	"math"
)

type Message struct {
	Identifier byte
	From       []byte
	Options    []byte
	Data       []byte
	Reply      chan Message
}

func NewMessage(id byte) *Message {
	return &Message{Identifier: id}
}

func (m *Message) MarshalBinary() ([]byte, error) {
	bs := &bytes.Buffer{}

	bs.WriteByte(m.Identifier)
	bs.Write(FitBytes(m.From, IP_SIZE))
	bs.Write(FitBytes(m.Options, MESSAGE_OPTIONS_SIZE))
	bs.Write(m.Data)
	return bs.Bytes(), nil
}

func (m *Message) UnMarshalBinary(d []byte) error {
	bs := bytes.NewBuffer(d)

	if len(d) < MESSAGE_OPTIONS_SIZE+MESSAGE_TYPE_SIZE {
		return errors.New("Insuficient message size")
	}
	m.Identifier = bs.Next(1)[0]
	m.From = bs.Next(IP_SIZE)
	m.Options = bs.Next(MESSAGE_OPTIONS_SIZE)
	m.Data = bs.Next(math.MaxInt64)

	return nil
}
