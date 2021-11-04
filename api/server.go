package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
	"github.com/mrod502/stockscraper/obj"
	"github.com/mrod502/stockscraper/scraper"
)

type Server struct {
	router      *mux.Router
	db          *db.DB
	v           *gocache.ObjectCache
	l           logger.Client
	s           scraper.Client
	c           Config
	newDocsChan chan *obj.Document
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
		err := d.Create()
		if err != nil {
			s.err("create", d.Source, err.Error())
			continue
		}
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
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)

	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var q db.DocQuery
	err = json.Unmarshal(b, &q)

	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := s.v.Where(q.Match)

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
