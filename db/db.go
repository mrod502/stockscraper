package db

import (
	"errors"

	badger "github.com/dgraph-io/badger/v3"
)

var (
	ErrClassNotFound       = errors.New("unable to determine class of db object")
	ErrUnknownType         = errors.New("unable to determine type of the object")
	ErrUnsupportedEncoding = errors.New("unsupported encoding")
)

type Typed interface {
	Type() uint16
}

type EncodingType int

const (
	Msgpack EncodingType = iota
	Json
	Bson
	Binary
)

type Config struct {
	BadgerOpts badger.Options `yaml:"badger_opts"`
	Encoding   EncodingType   `yaml:"encoding"`
	Compress   bool           `yaml:"compress"`
}

type DB struct {
	db  *badger.DB
	cfg Config
}

func New(cfg Config) (db *DB, err error) {
	d, err := badger.Open(cfg.BadgerOpts)

	if err != nil {
		return nil, err
	}

	db = &DB{
		db:  d,
		cfg: cfg,
	}
	return
}

func (d *DB) Close() error { return d.db.Close() }

func (d *DB) Get(typ uint16, k string) (*TypedObject, error) {
	return d.get(ToKey(k, typ))
}

func ToKey(k string, typ uint16) []byte {
	return append(append(TypeToBytes(typ), []byte("%")...), []byte(k)...)
}

func (d *DB) Put(k string, v Typed) error {
	t, err := FromTyped(v, d.cfg.Compress)
	if err != nil {
		return err
	}
	return d.put(ToKey(k, v.Type()), t.toBytes())
}

func (d *DB) Exists(typ uint16, k string) (bool, error) {
	if err := d.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(ToKey(k, typ))
		return err
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *DB) Delete(typ uint16, k string) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(ToKey(k, typ))
	})
}

func (d *DB) Keys(typ uint16, prefix string) (keys []string) {
	keys = make([]string, 0)
	d.db.View(func(t *badger.Txn) error {
		iterator := t.NewIterator(badger.DefaultIteratorOptions)
		defer iterator.Close()
		for iterator.Seek(ToKey(prefix, typ)); iterator.Valid(); iterator.Next() {
			key := iterator.Item().Key()
			keys = append(keys, string(key))
		}
		return nil
	})
	return
}
