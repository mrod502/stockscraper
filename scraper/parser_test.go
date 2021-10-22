package scraper

import (
	"fmt"
	"os"
	"testing"
)

func TestParser(t *testing.T) {

	p := &GParser{}

	b, _ := os.ReadFile("results_google.html")

	_, err := p.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	/*
		for _, doc := range docs {
			doc.Path = doc.Title + ".pdf"
			err := doc.Create()
			if err != nil {
				fmt.Println(err)
			}
		}
	*/
}

func TestStripQuery(t *testing.T) {

	uri := `https://www.krungsriasset.com/DataWeb/AYFWeb/en/pdf/Presentation_KFINNO-A_EN.pdf?rnd=20210602085041`
	s, err := stripQuery(uri)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}

func TestParseGoogle(t *testing.T) {
	if v, err := parseGoogleDate("Aug 23, 2020"); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(v)
	}
}
