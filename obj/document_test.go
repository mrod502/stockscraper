package obj

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

type TestFoo struct {
	Field1 int `msgpack:"Field1" json:"-"`
}

func TestDocument(t *testing.T) {
	f := TestFoo{
		Field1: 42,
	}

	b, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	b, err = msgpack.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

}
