package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	var cfg Config
	b, _ := json.Marshal(cfg)
	fmt.Println(string(b))
}
