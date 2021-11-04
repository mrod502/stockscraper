package scraper

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mrod502/stockscraper/obj"
)

func NewGoogleClient() *GoogleClient {
	return &GoogleClient{p: &GParser{}}
}

type GoogleClient struct {
	p *GParser
}

func (g *GoogleClient) Scrape(symbol, ftype string) (d []*obj.Document, err error) {

	uri := buildGoogleUri(fmt.Sprintf("%s equity research filetype:%s", symbol, ftype))
	r, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	setGoogleHeaders(r)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	d, err = g.p.Parse(b)
	for _, doc := range d {
		if doc != nil {
			doc.Symbols = []string{symbol}
		}
	}
	return
}
