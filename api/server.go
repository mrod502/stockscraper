package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
	"github.com/mrod502/stockscraper/obj"
	"github.com/mrod502/stockscraper/scraper"
	"go.uber.org/atomic"
)

type Server struct {
	router      *mux.Router
	db          *db.DB
	v           *gocache.ObjectCache
	l           logger.Client
	s           scraper.Client
	c           Config
	newDocsChan chan *obj.Document
	crawlReset  *atomic.Int64
}

func NewServer(cfg Config, errHandler func(error)) (s *Server, err error) {
	db, err := db.New(cfg.Db)
	if err != nil {
		return nil, err
	}
	l, err := logger.NewClient(cfg.Logger)
	if err != nil {
		return nil, err
	}
	s = &Server{
		router:      mux.NewRouter(),
		v:           gocache.NewObjectCache().WithDb(db),
		db:          db,
		newDocsChan: make(chan *obj.Document, 512),
		l:           l,
		c:           cfg,
		s:           scraper.NewGoogleClient(),
		crawlReset:  atomic.NewInt64(0),
	}
	s.l.SetLogLocally(true)
	err = s.l.Connect()
	if err != nil {
		return nil, err
	}
	s.buildRoutes()
	return
}
func (s *Server) Close() error { return s.db.Close() }

func (s *Server) documentProcessor() {
	for {
		d := <-s.newDocsChan
		fmt.Printf("processing document:\n\t%+v\n", *d)
		err := d.Create()
		if err != nil {
			fmt.Println("err", err.Error())
			s.err("create", d.Source, err.Error())
			continue
		}
		fmt.Println("putting", d.Id)
		s.db.Put(d.Id, d)
	}
}

func (s *Server) Serve() error {
	go s.documentProcessor()
	fmt.Println("listening on ", fmt.Sprintf(":%d", s.c.ServePort))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.c.ServePort), s.router)
}

func (s *Server) buildRoutes() {
	s.router.HandleFunc("/scrape/{symbol}/{filetype}", s.Scrape)
	s.router.HandleFunc("/query", s.Query)
	s.router.HandleFunc("/crawl", s.crawl)
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	b, err := io.ReadAll(r.Body)

	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var q db.DocQuery
	err = json.Unmarshal(b, &q)
	s.log("query", r.RemoteAddr, string(b))
	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := s.v.Where(q)

	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b, _ = json.Marshal(res)

	_, err = w.Write(b)
	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
	}

}

func (s *Server) Scrape(w http.ResponseWriter, r *http.Request) {
	enableCors(w)

	vars := mux.Vars(r)

	symbol := vars["symbol"]
	ftype := vars["filetype"]
	if !(ftype == "pdf" || ftype == "txt" || ftype == "html" || ftype == "xml") {
		http.Error(w, "invalid filetype", http.StatusBadRequest)
		s.err(requestSummary(r)...)
		return
	}
	go func() {
		d, err := s.s.Scrape(symbol, ftype)
		if err != nil {
			s.err("scrape", err.Error())
		}
		for _, v := range d {
			s.newDocsChan <- v
		}
	}()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) crawl(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if s.crawlReset.Load() > time.Now().Unix() {
		w.Header().Set("x-ratelimit-reset", fmt.Sprintf("%d", s.crawlReset.Load()))
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		s.err("crawl", r.RemoteAddr)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		s.err("crawl", r.RemoteAddr, err.Error())
		return
	}

	s.crawlReset.Store(time.Now().Unix() + 60)

	var params scraper.CrawlerParams
	err = json.Unmarshal(b, &params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		s.err("crawl", r.RemoteAddr, err.Error())
		return
	}

	c, err := scraper.NewCrawler(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.err("crawl", r.RemoteAddr, err.Error())
		return
	}

	c.SetApplicationFileHandler(func(r *http.Response) error {
		if r.Header.Get("content-type") == "application/pdf" {
			doc := obj.NewDocument(r)
			s.newDocsChan <- doc
		}
		return nil
	})

	c.SetTextFileHandler(func(r *http.Response) error {
		doc := obj.NewDocument(r)
		s.newDocsChan <- doc
		return nil
	})
	s.log("crawling")
	c.SetLogger(s.l)
	c.Crawl(8)

}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "privatekey")
	w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS,POST,HEAD,DELETE,PUT")
}
