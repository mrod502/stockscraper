package obj

import (
	"net/http"

	gocache "github.com/mrod502/go-cache"
)

type Provider interface {
	Provide(http.ResponseWriter) error
}

type Imitator interface {
	Get(*http.Request) (*http.Response, error)
	Post(*http.Request) (*http.Response, error)
}

type Searcher interface {
	Search(string) ([]gocache.Object, error)
}
