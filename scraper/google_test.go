package scraper

import (
	"io"
	"net/http"
	"os"
	"testing"
)

func TestGoogle(t *testing.T) {
	uri := `https://www.google.com/search?q=AMZN+financial+analysis+filetype%3Apdf&oq=AMZN+financial+analysis+filetype%3Apdf`
	var req *http.Request
	req, _ = http.NewRequest("GET", uri, nil)
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7")
	req.Header.Set("sec-ch-ua", `"Chromium";v="94", "Google Chrome";v="94", ";Not A Brand";v="99"`)
	req.Header.Set(`user-agent`, `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36`)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	f, _ := os.Create("results1.html")
	f.Write(b)
	f.Close()
}
