package scraper

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mrod502/stockscraper/obj"
)

type GoogleClient struct {
	p *GParser
}

func (g *GoogleClient) Scrape(symbol, ftype string) (d []*obj.Document, err error) {

	uri := buildGoogleUri(fmt.Sprintf("%s equity research filetype:%s", symbol, ftype))
	r, _ := http.NewRequest("GET", uri, nil)
	setGoogleHeaders(r)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return g.p.Parse(b)
}
