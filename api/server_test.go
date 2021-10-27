package api

import (
	"fmt"
	"os"
	"testing"
)

func TestPerm(t *testing.T) {
	info, err := os.Stat("config.go")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", info)
}
