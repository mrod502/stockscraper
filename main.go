package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrod502/stockscraper/api"
	"github.com/mrod502/stockscraper/obj"
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
	fmt.Printf("Config:\n%+v\n", cfg)

	r, err := api.NewServer(cfg, nil)
	if err != nil {
		panic(err)
	}
	obj.Setup(cfg.Obj)
	fmt.Println("listening")
	err = r.Serve()
	if err != nil {
		panic(err)
	}
}
