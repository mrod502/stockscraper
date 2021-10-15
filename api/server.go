package api

import (
	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/stockscraper/db"
)

type Server struct {
	router *mux.Router
	db     gocache.DB
	v      *gocache.ItemCache
}

func NewServer() (s *Server, err error) {
	var dbOpts db.Options

	db, err := db.New(dbOpts)
	if err != nil {
		return nil, err
	}
	s = &Server{
		router: mux.NewRouter(),
		v:      gocache.NewItemCache().WithDb(db),
		db:     db,
	}

	return
}
