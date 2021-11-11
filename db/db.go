package db

import (
	"bytes"
	"errors"

	badger "github.com/dgraph-io/badger/v3"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/stockscraper/obj"
	msgpack "github.com/vmihailenco/msgpack/v5"
)

var (
	ErrClassNotFound = errors.New("unable to determine class of db object")
	ErrUnknownType   = errors.New("unable to determine type of the object")
)

type Config struct {
	BadgerOpts badger.Options `yaml:"badger_opts"`
}

type DB struct {
	db *badger.DB
}

func New(cfg Config) (db *DB, err error) {

	d, err := badger.Open(cfg.BadgerOpts)

	if err != nil {
		return nil, err
	}
	db = &DB{
		db: d,
	}
	return
}
func (d *DB) Close() error { return d.db.Close() }

func (d *DB) Get(k string) (gocache.Object, error) {
	var obj gocache.Object
	d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			obj, err = d.getItemObject(val)
			return err
		})
		return err
	})
	return obj, nil
}

func (d *DB) Put(k string, v gocache.Object) error {
	return d.db.Update(func(txn *badger.Txn) error {
		b, err := msgpack.Marshal(v)
		if err != nil {
			return err
		}
		if err = txn.Set([]byte(k), b); err != nil {
			return err
		}
		return txn.Commit()
	})
}

func (d *DB) Exists(k string) (bool, error) {
	if err := d.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(k))
		return err
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *DB) Delete(k string) error {

	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(k))
	})
}

func (d *DB) Keys() (keys []string) {
	keys = make([]string, 0)
	d.db.View(func(t *badger.Txn) error {
		iterator := t.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		for iterator.Seek([]byte("")); iterator.Valid(); iterator.Next() {
			key := iterator.Item().Key()
			keys = append(keys, string(key))
		}
		return nil
	})
	return
}

func GetClass(b []byte) (string, error) {
	ix0 := bytes.Index(b, []byte("Class")) + 6
	if ix0 < 0 {
		return "", ErrClassNotFound
	}
	ix1 := bytes.Index(b[ix0:], []byte{168}) + ix0
	return string(b[ix0:ix1]), nil
}

func BytesToObj(b []byte) (interface{}, error) {
	c, err := GetClass(b)
	if err != nil {
		return nil, err
	}
	switch c {
	case obj.TDocument:
		var v obj.Document
		err := msgpack.Unmarshal(b, &v)
		return v, err
	default:
		return nil, ErrUnknownType
	}
}

func (d *DB) getItemObject(b []byte) (gocache.Object, error) {
	t, err := GetClass(b)
	if err != nil {
		return nil, err
	}
	switch t {
	case obj.TDocument:
		var obj = &obj.Document{}
		err := msgpack.Unmarshal(b, obj)
		return obj, err
	default:
		return nil, ErrUnknownType
	}

}
