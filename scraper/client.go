package scraper

import (
	"fmt"
	"strings"

	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/stockscraper/obj"
)

const (
	googleSearch = "https://www.google.com/search"
)

const example = `https://www.google.com/search?q=AMZN+financial+analysis+filetype%3Apdf&oq=AMZN+financial+analysis+filetype%3Apdf&aqs=chrome..69i57.11722j1j7&sourceid=chrome&ie=UTF-8`

func buildGoogleUri(q string) (u string) {
	q = strings.ReplaceAll(q, " ", "+")
	q = strings.ReplaceAll(q, ":", `%3A`)
	return fmt.Sprintf(googleSearch+"?q=%s&oq=%s&sourceid=chrome&ie=UTF-8", q, q)
}

func buildBingUri(q string) (u string) {
	q = strings.ReplaceAll(q, " ", "+")
	q = strings.ReplaceAll(q, ":", `%3A`)
	return fmt.Sprintf(googleSearch+"?q=%s&oq=%s&sc=0-33&qs=n", q, q)
}

type Client struct {
	b obj.Imitator
}

func (c *Client) ScrapeSymbol(s string) (d []obj.Document, err error) {

	return
}

func (c *Client) GetCompanyName(s string) (v string, err error) {

	return
}

type ChromeImitator struct {
	jar *gocache.ItemCache
}
