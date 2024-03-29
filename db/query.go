package db

import (
	"github.com/dgraph-io/badger/v3"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/stockscraper/obj"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	DefaultResultLimit uint = 100
)

type Query[T any] interface {
	Prefix() string
	Unmarshal([]byte, func([]byte, interface{}) error) (T, error)
	Match(T) bool
}

func (d *DB) Where(m Query[any]) ([]any, error) {
	var matches uint
	for _, key := range d.Keys() {
		err := d.db.View(func(t *badger.Txn) error {
			item, rerr := t.Get([]byte(key))
			if rerr != nil {
				return rerr
			}
			return item.Value(func(b []byte) error {
				class, err := GetClass(b)
				if err != nil {
					return err
				}
				switch class {
				case obj.TDocument:
					var object = new(obj.Document)
					err = msgpack.Unmarshal(b, object)
					if err != nil {
						return err
					}
					if m.Match(object) {
						matches++
						objects = append(objects, object)
					}
				default:
					return ErrClassNotFound
				}
				return err
			})
		})
		if err != nil {
			return objects, err
		}
		if matches >= m.GetLimit() {
			return objects, nil
		}
	}
	return objects, nil
}

type ItemQuery struct {
	Created  gocache.TimeQuery
	Class    gocache.StringQuery
	Archived gocache.BoolQuery
	Limit    uint
}

func (q ItemQuery) GetLimit() uint { return 10000 }

func NewItemQuery(c gocache.TimeQuery, cl gocache.StringQuery, d gocache.BoolQuery, l uint) ItemQuery {

	return ItemQuery{
		Created:  c,
		Class:    cl,
		Archived: d,
		Limit:    l,
	}
}
func NewDocQuery(i ItemQuery,
	tit gocache.StringQuery,
	sym gocache.StringQuery,
	sec gocache.StringQuery,
	src gocache.StringQuery,
	ctp gocache.StringQuery,
	typ gocache.StringQuery,
	pdt gocache.TimeQuery) DocQuery {

	return DocQuery{
		ItemQuery:   i,
		Title:       tit,
		Symbols:     sym,
		Sectors:     sec,
		Source:      src,
		ContentType: ctp,
		Type:        typ,
		PostedDate:  pdt,
	}
}

func (i ItemQuery) Match(v gocache.Object) bool {
	item := v.(obj.Item)

	return i.Created.Match(item.Created) && i.Class.Match(item.Class) && i.Archived.Match(item.Archived)
}

type DocQuery struct {
	ItemQuery
	Title       gocache.StringQuery
	Symbols     gocache.StringQuery
	Sectors     gocache.StringQuery
	Source      gocache.StringQuery
	ContentType gocache.StringQuery
	Type        gocache.StringQuery
	PostedDate  gocache.TimeQuery
}

func (d DocQuery) Match(v gocache.Object) bool {
	doc := v.(*obj.Document)
	return d.ItemQuery.Match(*doc.Item) && d.Title.Match(doc.Title) &&
		d.Symbols.Match(doc.Symbols) && d.Sectors.Match(doc.Sectors) &&
		d.Source.Match(doc.Source) && d.ContentType.Match(doc.ContentType) &&
		d.Type.Match(doc.ContentType) && d.PostedDate.Match(doc.PostedDate)

}

func (d DocQuery) GetLimit() uint {
	return ifZero(d.Limit, DefaultResultLimit)
}
