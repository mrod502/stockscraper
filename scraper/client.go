package scraper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mrod502/stockscraper/obj"
)

const (
	googleSearch = "https://www.google.com/search"
)

type Config struct {
}

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

type Client interface {
	Scrape(symbol, filetype string) (chan *obj.Document, error)
}

func setGoogleHeaders(req *http.Request) {
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7")
	req.Header.Set("sec-ch-ua", `"Chromium";v="94", "Google Chrome";v="94", ";Not A Brand";v="99"`)
	req.Header.Set(`user-agent`, `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36`)
}
