package obj

import (
	"net/http"
)

type Provider interface {
	Provide(http.ResponseWriter) error
}

type Imitator interface {
	Get(*http.Request) (*http.Response, error)
	Post(*http.Request) (*http.Response, error)
}
