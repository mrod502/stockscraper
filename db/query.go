package db

import (
	badger "github.com/dgraph-io/badger/v3"
	gocache "github.com/mrod502/go-cache"
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

func (d *DB) Each(f func(TypedObject) error, prefix string, limit uint) error {
	d.db.View(func(t *badger.Txn) error {
		iterator := t.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		for iterator.Seek([]byte(prefix)); iterator.Valid(); iterator.Next() {
			obj, err := d.load(iterator.Item())
			if err != nil {
				return err
			}
			return f(*obj)
		}
		return nil
	})
	return nil
}

type ItemQuery struct {
	Created  gocache.TimeQuery
	Class    gocache.StringQuery
	Archived gocache.BoolQuery
	Limit    uint
}

func (q ItemQuery) GetLimit() uint { return 10000 }

func (i ItemQuery) Match(v any) bool {
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
