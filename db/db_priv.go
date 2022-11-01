package db

import badger "github.com/dgraph-io/badger/v3"

func (d *DB) put(k []byte, v []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
}

func (d *DB) get(k []byte) (*TypedObject, error) {
	var obj *TypedObject
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		obj, err = d.load(item)
		return err
	})
	return obj, err
}

func (d *DB) load(item *badger.Item) (*TypedObject, error) {
	b, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	var obj = new(TypedObject)
	err = obj.fromBytes(b)
	return obj, err
}
