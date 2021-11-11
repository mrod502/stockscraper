package scraper

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ScanFile(src string, conditions ...func(string)) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}

	var buf = make([]byte, 512)

	_, err = f.Read(buf)
	if err != nil {
		return err
	}
	ct := http.DetectContentType(buf)
	var reader io.Reader
	switch ct {
	case "text/plain", "text/html", "text/xml":
		reader = bufio.NewReader(f)
	case "application/x-gzip":
		reader, err = gzip.NewReader(f)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unable to process file %s of type %s", src, ct)
	}

	return scanTextFile(bufio.NewScanner(reader), conditions...)

}

func scanTextFile(f *bufio.Scanner, conditions ...func(string)) error {
	for f.Scan() {
		t := f.Text()
		for _, cond := range conditions {
			cond(t)
		}
	}
	return nil

}
