package db

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	Id       [16]byte
	Created  time.Time
	Class    string
	Archived bool `msgpack:"md"`
}

func NewItem(class string) *Item {
	return &Item{
		Id:      uuid.New(),
		Created: time.Now(),
		Class:   class,
	}
}
