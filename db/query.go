package db

import (
	"bytes"
	"errors"
	"regexp"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/mrod502/stockscraper/obj"
)

var (
	DefaultResultLimit uint = 100
)

type Query[T any] interface {
	Prefix() string
	Unmarshal([]byte, func([]byte, interface{}) error) (T, error)
	Match(T) bool
}

func (d *DB) Each(f func(TypedObject) error, prefix string, typ uint16) error {
	d.db.View(func(t *badger.Txn) error {
		iterator := t.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		for iterator.Seek(ToKey(prefix, typ)); iterator.Valid(); iterator.Next() {
			obj, err := d.load(iterator.Item())
			if err != nil {
				return err
			}
			if err := f(*obj); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

type ItemQuery struct {
	Created  TimeQuery
	Class    StringQuery
	Archived BoolQuery
	Limit    uint
}

func (q ItemQuery) GetLimit() uint { return 10000 }

func (i ItemQuery) Match(v any) bool {
	item := v.(obj.Item)

	return i.Created.Match(item.Created) && i.Class.Match(item.Class) && i.Archived.Match(item.Archived)
}

type DocQuery struct {
	ItemQuery
	Title       StringQuery
	Symbols     StringQuery
	Sectors     StringQuery
	Source      StringQuery
	ContentType StringQuery
	Type        StringQuery
	PostedDate  TimeQuery
}

func (d DocQuery) Match(v any) bool {
	doc := v.(*obj.Document)
	return d.ItemQuery.Match(*doc.Item) && d.Title.Match(doc.Title) &&
		d.Symbols.Match(doc.Symbols) && d.Sectors.Match(doc.Sectors) &&
		d.Source.Match(doc.Source) && d.ContentType.Match(doc.ContentType) &&
		d.Type.Match(doc.ContentType) && d.PostedDate.Match(doc.PostedDate)

}

func (d DocQuery) GetLimit() uint {
	return ifZero(d.Limit, DefaultResultLimit)
}

var (
	ErrInterfaceAssertion error = errors.New("interface type-assertion failed")
)

const (
	Greater byte = iota + 1
	Less
	GreaterEq
	LessEq
	Eq
	Neq
	Regex
)

func NewTimeQuery(v time.Time, c byte, rex string) TimeQuery {
	return TimeQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type TimeQuery struct {
	V     time.Time
	C     byte
	Check bool `msgpack:"chk"`
}

func (q TimeQuery) Match(i interface{}) bool {
	v, ok := i.(time.Time)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v.UnixNano() > q.V.UnixNano()
	case GreaterEq:
		return v.UnixNano() >= q.V.UnixNano()
	case LessEq:
		return v.UnixNano() <= q.V.UnixNano()
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewIntQuery(v int, c byte, rex string) IntQuery {
	return IntQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type IntQuery struct {
	V     int
	C     byte
	Check bool `msgpack:"chk"`
}

func (q IntQuery) Match(i interface{}) bool {
	if !q.Check {
		return true
	}

	switch val := i.(type) {
	case int:
		return q.cmp(val)
	case []int:
		for _, v := range val {
			if q.cmp(v) {
				return true
			}
		}
		return false
	default:
		return false
	}

}

func (q IntQuery) cmp(v int) bool {
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewFloatQuery(v float64, c byte, rex string) FloatQuery {
	return FloatQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type FloatQuery struct {
	V     float64
	C     byte
	Check bool `msgpack:"chk"`
}

func (q FloatQuery) Match(i interface{}) bool {
	v, ok := i.(float64)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewStringQuery(v string, c byte, rex string) StringQuery {
	return StringQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type StringQuery struct {
	V     string
	C     byte
	Check bool `msgpack:"chk"`
}

func (q StringQuery) Match(i interface{}) bool {
	if !q.Check {
		return true
	}
	switch v := i.(type) {
	case string:
		return q.match(v)
	case []string:
		for _, val := range v {
			if q.match(val) {
				return true
			}
		}
	}
	return false
}

func (q StringQuery) match(v string) bool {
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Regex:
		if len(q.V) == 0 {
			return q.V == v
		}
		rex, err := regexp.Compile(q.V)
		if err != nil {
			return false
		}
		return len(rex.FindString(v)) > 0
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewByteQuery(v byte, c byte) ByteQuery {
	return ByteQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type ByteQuery struct {
	V     byte
	C     byte
	Check bool `msgpack:"chk"`
}

func (q ByteQuery) Match(i interface{}) bool {
	v, ok := i.(byte)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewByteSliceQuery(v []byte, c byte, rex string) ByteSliceQuery {
	return ByteSliceQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type ByteSliceQuery struct {
	V     []byte
	C     byte
	Check bool `msgpack:"chk"`
}

func (q ByteSliceQuery) Match(i interface{}) bool {
	if !q.Check {
		return true
	}
	switch val := i.(type) {
	case []byte:
		return q.match(val)
	case [][]byte:
		for _, v := range val {
			if q.match(v) {
				return true
			}
		}
	}
	return false
}

func (q ByteSliceQuery) match(v []byte) bool {
	switch q.C {
	case Greater:
		return string(v) > string(q.V)
	case GreaterEq:
		return string(v) >= string(q.V)
	case LessEq:
		return string(v) <= string(q.V)
	case Regex:
		if len(q.V) == 0 {
			return bytes.Equal(q.V, v)
		}
		rex, err := regexp.Compile(string(q.V))
		if err != nil {
			return false
		}
		return len(rex.Find(v)) > 0
	case Neq:
		return !bytes.Equal(q.V, v)
	default:
		return bytes.Equal(q.V, v)
	}
}

func NewBoolQuery(v bool, c byte) BoolQuery {
	return BoolQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type BoolQuery struct {
	V     bool
	C     byte
	Check bool `msgpack:"chk"`
}

func (q BoolQuery) Match(i interface{}) bool {
	v, ok := i.(bool)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v && !q.V
	case GreaterEq:
		return v
	case LessEq:
		return !v
	case Neq:
		return v != q.V
	case Eq:
		return v == q.V
	default:
		return v
	}
}

func Or(v ...bool) bool {
	for _, val := range v {
		if val {
			return true
		}
	}
	return false
}
func And(v ...bool) bool {
	for _, val := range v {
		if !val {
			return false
		}
	}
	return true
}
