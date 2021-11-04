package db

import (
	"github.com/dgraph-io/badger/v3"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/stockscraper/obj"
	"github.com/vmihailenco/msgpack/v5"
)

func (d *DB) Where(m gocache.Matcher) ([]gocache.Object, error) {
	var objects []gocache.Object = make([]gocache.Object, 0)
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
					if m(object) {
						objects = append(objects, object)
					}
				default:
					return ErrClassNotFound

				}
				return err
			})

		})
		if err != nil {
			return nil, err
		}

	}
	return nil, nil
}

type ItemQuery struct {
	Created  gocache.TimeQuery
	Class    gocache.StringQuery
	Archived gocache.BoolQuery
}

func NewItemQuery(c gocache.TimeQuery, cl gocache.StringQuery, d gocache.BoolQuery) ItemQuery {

	return ItemQuery{
		Created:  c,
		Class:    cl,
		Archived: d,
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

	return d.ItemQuery.Match(doc.Item) && d.Title.Match(doc.Title) &&
		d.Symbols.Match(doc.Symbols) && d.Sectors.Match(doc.Sectors) &&
		d.Source.Match(doc.Source) && d.ContentType.Match(doc.ContentType) &&
		d.Type.Match(doc.ContentType) && d.PostedDate.Match(doc.PostedDate)

}
