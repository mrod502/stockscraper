package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/mrod502/stockscraper/api"
	"github.com/mrod502/stockscraper/obj/document"
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
	document.Setup(cfg.Obj)
	defer r.Close()
	go func() {
		err = r.Serve()
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("bye")
}
