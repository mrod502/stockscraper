package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
	"github.com/mrod502/stockscraper/scraper"
)

type Server struct {
	router       *mux.Router
	db           *db.DB
	v            *gocache.ObjectCache
	l            logger.Client
	s            scraper.Client
	c            Config
	errorHandler func(error)
}

func NewServer(cfg Config, errHandler func(error)) (s *Server, err error) {

	db, err := db.New(cfg.Db)
	if err != nil {
		return nil, err
	}
	s = &Server{
		router: mux.NewRouter(),
		v:      gocache.NewObjectCache().WithDb(db),
		db:     db,
	}
	if errHandler == nil {
		s.errorHandler = s.defaultErrorHandler
	}
	s.buildRoutes()
	return
}

func (s *Server) defaultErrorHandler(err error) {
	s.l.Write("API", "scraper", err.Error())
}

func (s *Server) Serve() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.c.ServePort), s.router)
}

func (s *Server) buildRoutes() {
	s.router.HandleFunc("/scrape/{symbol}/{filetype}", s.Scrape)
	s.router.HandleFunc("/query", s.Query)
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) {

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
	go s.s.Scrape(symbol, ftype)
	w.WriteHeader(http.StatusOK)
}
