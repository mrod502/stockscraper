package main

import (
	"encoding/json"
	"os"

	"github.com/mrod502/stockscraper/api"
)

func main() {
	var cfg api.Config

	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}

	r, err := api.NewServer(cfg, nil)
	if err != nil {
		panic(err)
	}
	err = r.Serve()
	if err != nil {
		panic(err)
	}
}
