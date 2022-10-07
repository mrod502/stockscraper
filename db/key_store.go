package db

import "sync"

type keyStore struct {
	sync.RWMutex
	v        string
	children []*keyStore
}

func (k *keyStore) Prefix(p, v string, cur []string) {
	if len(k.children) > 0 {
		for _, c := range k.children {
			c.Prefix(p, v, cur)
		}
	}
}

type KeyStore struct {
	sync.RWMutex
	keys *keyStore
}

func (k *keyStore) Keys(prefix string) []string {
	return make([]string, 0)
}
