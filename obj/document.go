package obj

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNoExtension   = errors.New("no file extension")
	ErrFilenameParse = errors.New("unable to parse uri into filename")
	ErrFileType      = errors.New("incorrect filetype")
	rexUrlFile       = regexp.MustCompile(`/([^/]+)$`)
)

func NewDocument(dbRoot, src string, r *http.Response) (d *Document, err error) {
	if err != nil {
		return nil, err
	}
	d = &Document{
		Item:        NewItem(TDocument),
		Source:      r.Request.URL.EscapedPath(),
		ContentType: r.Header.Get("content-type"),
	}
	return
}

type Document struct {
	*Item
	Title       string    `msgpack:"tit,omitempty"`
	Symbols     []string  `msgpack:"sym,omitempty"` // Stock or crypto symbols mentioned in the article
	Sectors     []string  `msgpack:"sct,omitempty"` // Sectors of industry / finance this document mentions / relates to
	Source      string    `msgpack:"src,omitempty"` // The URL of this document
	ContentType string    `msgpack:"ctt,omitempty"` // the content type (pdf,etc)
	Type        string    `msgpack:"typ,omitempty"` // Financial statement, analysis, blog post, etc...
	PostedDate  time.Time `msgpack:"pdate,omitempty"`
}

func getFilenameFromUri(url string) (string, error) {
	match := rexUrlFile.FindStringSubmatch(url)
	if len(match) < 2 {
		return "", ErrFilenameParse
	}
	return match[1], nil
}

func (d *Document) Create() error {
	return docMgr.save(d)
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

func Includes(s string, v []string) bool {
	for _, val := range v {
		if strings.Contains(val, s) {
			return true
		}
	}
	return false
}

func (d *Document) retrieve() ([]byte, error) {
	req := generateBrowserRequest(d.Source)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	d.ContentType = res.Header.Get("content-type")
	return io.ReadAll(res.Body)
}

func generateBrowserRequest(uri string) *http.Request {
	var req *http.Request
	req, _ = http.NewRequest("GET", uri, nil)
	req.Header.Set("accept", "application/pdf,text/html,text/plain,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7")
	req.Header.Set("sec-ch-ua", `"Chromium";v="94", "Google Chrome";v="94", ";Not A Brand";v="99"`)
	req.Header.Set(`user-agent`, `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36`)
	return req
}
