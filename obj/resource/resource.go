package obj

import (
	"bufio"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type Resource struct {
	FilePath    string `json:"-" msgpack:"fpath"`
	ContentType string `json:"-" msgpack:"ctype"`
	IsGZipped   bool   `json:"-" msgpack:"gz"`
	Source      string `msgpack:"src"`
}

func (r Resource) LoadContent(basepath ...string) ([]byte, error) {
	reader, err := r.Load(basepath...)
	if err != nil {
		return nil, err
	}
	var b = make([]byte, 0)
	_, err = reader.Read(b)

	return b, err
}

func (r Resource) Load(basePath ...string) (io.Reader, error) {
	fpath := r.FilePath
	if len(basePath) > 0 {
		fpath = path.Join(basePath[0], fpath)
	}

	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	if !r.IsGZipped {
		return f, nil
	}

	gzreader, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(gzreader), nil
}

func FromResponse(res *http.Response) (r *Resource, err error) {
	fileName := ""
	r = &Resource{
		ContentType: res.Header.Get("content-type"),
		IsGZipped:   strings.Contains(fileName, ".gz"),
		Source:      res.Request.URL.EscapedPath(),
	}

	return
}
