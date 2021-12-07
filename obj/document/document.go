package document

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mrod502/stockscraper/obj/item"
	"github.com/mrod502/stockscraper/obj/types"
)

var (
	ErrNoExtension   = errors.New("no file extension")
	ErrFilenameParse = errors.New("unable to parse uri into filename")
	ErrFileType      = errors.New("incorrect filetype")
)
var (
	rexSrc = regexp.MustCompile(`src='([^']+)'`)
)

func New(r *http.Response) (d *Document) {
	d = &Document{
		Item:        item.New(types.Document),
		Source:      "https://" + r.Request.Host + r.Request.URL.RequestURI(),
		ContentType: r.Header.Get("content-type"),
	}
	return
}

type Document struct {
	*item.Item
	Title       string    `msgpack:"tit,omitempty"`
	Symbols     []string  `msgpack:"sym,omitempty"` // Stock or crypto symbols mentioned in the article
	Sectors     []string  `msgpack:"sct,omitempty"` // Sectors of industry / finance this document mentions / relates to
	Source      string    `msgpack:"src,omitempty"` // The URL of this document
	ContentType string    `msgpack:"ctt,omitempty"` // the content type (pdf,etc)
	Type        string    `msgpack:"typ,omitempty"` // Financial statement, analysis, blog post, etc...
	PostedDate  time.Time `msgpack:"pdate,omitempty"`
}

func (d *Document) Create() error {
	docMgr.saveChan <- d
	return nil
}

func (d *Document) Destroy() error {
	return docMgr.remove(d)
}

func (d *Document) Provide(w http.ResponseWriter) error {
	b, err := docMgr.load(d)
	if err != nil {
		return err
	}
	w.Header().Set("content-type", d.ContentType)
	_, err = w.Write(b)
	return err
}

func (d *Document) Load() (*os.File, error) {
	return docMgr.loadFile(d)
}

func Includes(s string, v []string) bool {
	for _, val := range v {
		if strings.Contains(val, s) {
			return true
		}
	}
	return false
}

func (d *Document) doRequest() (res *http.Response, err error) {
	req := generateBrowserRequest(d.Source)
	if req == nil {
		return nil, errors.New("nil request")
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	d.ContentType = res.Header.Get("content-type")
	return
}

func (d *Document) retrieve() ([]byte, error) {
	res, err := d.doRequest()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println("doc: got - ", res.Request.URL.EscapedPath())
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if d.ContentType == "text/html" {
		if m := rexSrc.FindStringSubmatch(string(b)); len(m) == 2 {
			if strings.Contains(m[1], "pdf") {
				d.Source = m[1]
				res, err := d.doRequest()
				if err != nil {
					return nil, err
				}
				return io.ReadAll(res.Body)
			}
		}

	}
	return b, err
}

func generateBrowserRequest(uri string) *http.Request {
	var req *http.Request
	var err error
	req, err = http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Println("generateBrowserRequest", uri, err)
		return req
	}
	req.Header.Set("accept", "application/pdf,text/html,text/plain,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7")
	req.Header.Set("sec-ch-ua", `"Chromium";v="94", "Google Chrome";v="94", ";Not A Brand";v="99"`)
	req.Header.Set(`user-agent`, `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36`)
	return req
}

func (d Document) isText() bool {
	return strings.Contains(d.ContentType, "text/")
}
