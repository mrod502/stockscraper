package db

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestDocument(t *testing.T) {
	uri := `https://www8.gsb.columbia.edu/valueinvesting/sites/valueinvesting/files/ASML%20NV%20Stock%20Pitch%20-%20sm4843%20-%20VI%20with%20Legends%20-%20final.pdf`
	res, err := http.Get(uri)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create("foo.pdf")
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	n, err := f.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(n)
	os.Remove("foo.pdf")

	fmt.Printf("%+v\n", res.Header)
}
