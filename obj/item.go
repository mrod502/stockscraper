package obj

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	Id       string
	Created  time.Time
	Class    string
	Archived bool
}

func NewItem(class string) *Item {
	return &Item{
		Id:      uuid.New().String(),
		Created: time.Now(),
		Class:   class,
	}
}

func (i Item) Create() error  { return nil }
func (i Item) Destroy() error { return nil }
