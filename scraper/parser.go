package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/mrod502/stockscraper/obj"
	"golang.org/x/net/html"
)

var (
	ErrFileName = errors.New("file name not found")
)

var (
	ftypeRex    = regexp.MustCompile(`\.([a-zA-Z0-9]+)$`)
	gDateRex    = regexp.MustCompile(`[a-zA-Z]+ [\d]+, [\d]+`)
	gYearRex    = regexp.MustCompile(`[\d]{4}`)
	fileNameRex = regexp.MustCompile(`/([^/]+)$`)
)

type GParser struct {
}

func (p *GParser) Parse(b []byte) (d []*obj.Document, err error) {
	defer func() {
		for _, v := range d {
			fmt.Printf("%+v\n", v)
		}
	}()
	d = make([]*obj.Document, 0, 5)
	z := html.NewTokenizer(bytes.NewReader(b))
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			for i, v := range d {
				if v.Source == "" {
					if i == (len(d) - 1) {
						d = d[:i]
					}
					d = append(d[:i], d[i+1:]...)
				}
			}
			return
		}
		switch tt {
		case html.StartTagToken:
			t := z.Token()
			attrs := t.Attr
			for _, a := range attrs {
				if a.Key == "class" && a.Val == "g" {
					doc := p.parseSearchResult(z)
					d = append(d, doc)
					break
				}
			}
		default:
		}
	}
}

func (p *GParser) parseSearchResult(t *html.Tokenizer) *obj.Document {

	level := 0
	var doc = new(obj.Document)
	doc.Item = obj.NewItem(obj.TDocument)

	for {
		tt := t.Next()
		switch tt {
		case html.StartTagToken:
			level++
			tk := t.Token()
			processToken(t, tk, doc)
		case html.EndTagToken:
			level--
		}
		if level == 0 {
			break
		}
	}
	return doc
}

func processToken(t *html.Tokenizer, tk html.Token, doc *obj.Document) error {
	d := tk.Data
	switch d {
	case "a":
		addSource(tk, doc)
	case "h3":
		addTitle(t, doc)
	case "span":
		handleSpan(tk, t, doc)
	default:
	}
	return nil
}

func handleSpan(tk html.Token, t *html.Tokenizer, doc *obj.Document) {
	// prob a better way to do this

	if len(tk.Attr) == 1 && tk.Attr[0].Val == "MUxGbd wuQ4Ob WZ8Tjf" {
		t.Next()
		tk = t.Token()
		if d, err := parseGoogleDate(tk.Data); err == nil {
			doc.PostedDate = d
		} else {
			doc.PostedDate, _ = time.Parse("2006", gYearRex.FindString(tk.Data))
		}
	}
}

func addTitle(t *html.Tokenizer, doc *obj.Document) {
	t.Next()
	data := t.Token().Data
	doc.Title = data
}

func addSource(tk html.Token, doc *obj.Document) {
	if href, has := getHref(tk.Attr); has {
		if v, err := stripQuery(href); err == nil {
			doc.Source = v
			doc.ContentType = getLinkContentType(v)
		}
	}
}
func getHref(attrs []html.Attribute) (string, bool) {
	for _, v := range attrs {
		if v.Key == "href" {
			return v.Val, true
		}
	}
	return "", false
}

func getLinkContentType(l string) string {
	if ftype := ftypeRex.FindStringSubmatch(l); len(ftype) > 1 {
		return "application/" + ftype[1]
	}
	return ""
}

func stripQuery(in string) (string, error) {
	uri, err := url.Parse(in)
	if err != nil {
		return "", err
	}
	return uri.Scheme + "://" + uri.Hostname() + uri.EscapedPath(), nil
}

func parseGoogleDate(s string) (time.Time, error) {
	return time.Parse("Jan 2, 2006", gDateRex.FindString(s))
}

func parseFileName(inp string) (string, error) {
	if matches := fileNameRex.FindStringSubmatch(inp); len(matches) > 1 {
		return matches[1], nil
	}
	return "", ErrFileName
}
