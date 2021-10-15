package db

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	ErrNoExtension = errors.New("no file extension")
)

func NewDocument(src string, r *http.Response) (d *Document, err error) {
	d = &Document{
		Item:        NewItem("Document"),
		Source:      r.Request.URL.EscapedPath(),
		ContentType: strings.ReplaceAll(strings.ReplaceAll(r.Header.Get("content-type"), "application/", ""), "text/", ""),
	}

	return
}

type Document struct {
	*Item
	Title       string   `msgpack:"tit,omitempty"`
	Symbols     []string `msgpack:"sym,omitempty"` // Stock or crypto symbols mentioned in the article
	Sectors     []string `msgpack:"sct,omitempty"` // Sectors of industry / finance this document mentions / relates to
	Source      string   `msgpack:"src,omitempty"` // The URL of this document
	ContentType string   `msgpack:"ctt,omitempty"` // the content type (pdf,etc)
	Type        string   `msgpack:"typ,omitempty"` // Financial statement, analysis, blog post, etc...
	Path        string   `msgpack:"pth,omitempty"` // Absolute (or rel) filepath
}

func (d *Document) Create() error {

	fileBytes, err := resourceToBytes(d.Source)

	if err != nil {
		return err
	}
	f, err := os.Create(d.Path)
	if err != nil {
		return err
	}
	_, err = f.Write(fileBytes)
	if err != nil {
		return err
	}

	return nil
}

func (d *Document) Destroy() error {

	return nil
}

func (d *Document) Provide(w http.ResponseWriter) error {

	return nil
}

func Includes(s string, v []string) bool {
	for _, val := range v {
		if strings.Contains(val, s) {
			return true
		}
	}
	return false
}

func (d *Document) FileType() (string, error) {
	if d.ContentType != "" {
		return d.ContentType, nil
	}
	v := strings.Split(d.Path, ".")
	if len(v) < 2 {
		return "", ErrNoExtension
	}
	d.ContentType = v[len(v)-1]
	return v[len(v)-1], nil
}

func resourceToBytes(uri string) ([]byte, error) {
	req := generateBrowserRequest(uri)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

func generateBrowserRequest(uri string) *http.Request {
	var req *http.Request
	req, _ = http.NewRequest("GET", uri, nil)

	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7")
	req.Header.Set("sec-ch-ua", `"Chromium";v="94", "Google Chrome";v="94", ";Not A Brand";v="99"`)
	return req
}