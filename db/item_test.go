package db

import (
	"fmt"
	"testing"

	"github.com/mrod502/stockscraper/obj"
	"github.com/vmihailenco/msgpack/v5"
)

type TestObj struct {
	*obj.Item
	Val       string
	Something struct {
		Val int
	}
}

func TestItem(t *testing.T) {
	var v = &TestObj{
		Item:      obj.NewItem("TestObj"),
		Something: struct{ Val int }{Val: 1},
	}
	b, err := msgpack.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("bytes:", b)
	c, _ := GetClass(b)
	fmt.Println(c)

	fmt.Println("str:", string(b))
}
