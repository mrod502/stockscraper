package db

import (
	"errors"

	"github.com/vmihailenco/msgpack/v5"
)

type TypedObject struct {
	Type uint16
	Data []byte
}

func FromTyped(t Typed, useCompression bool) (TypedObject, error) {
	b, err := msgpack.Marshal(t)
	if err != nil {
		return TypedObject{}, err
	}

	return TypedObject{
		Type: t.Type(),
		Data: b,
	}, nil
}

func (t TypedObject) toBytes() []byte {
	var typ = []byte{byte(t.Type >> 8), byte(t.Type)}
	return append(typ, t.Data...)
}

func (t *TypedObject) fromBytes(b []byte) error {
	if len(b) < 2 {
		return errors.New("b not long enough")
	}
	t.Type = (uint16(b[0]) << 8) + uint16(b[1])
	if len(b) > 2 {
		t.Data = b[2:]
	} else {
		t.Data = make([]byte, 0)
	}
	return nil
}
