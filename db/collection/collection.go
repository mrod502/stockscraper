package collection

import (
	"github.com/dgraph-io/badger/v3"
)

type Collection struct {
	coll *badger.DB
	name string
}

func New(opts badger.Options, name string) (coll *Collection, err error) {

	return
}
