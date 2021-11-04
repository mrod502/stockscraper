package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

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
	defer r.Close()
	go func() {
		err = r.Serve()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		b, err := os.ReadFile("symbols.txt")
		if err != nil {
			fmt.Println(err)
			return
		}
		s := string(b)
		for _, sym := range strings.Split(s, "\n") {
			http.DefaultClient.Get(fmt.Sprintf("http://localhost:8849/scrape/%s/pdf", sym))
			time.Sleep(time.Second * time.Duration(rand.Intn(40)))
		}

	}()
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("bye")
}
