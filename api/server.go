package api

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
)

type scrapeAction struct {
}

type Server struct {
	router       *mux.Router
	db           gocache.DB
	v            *gocache.ItemCache
	scrapeQueue  chan scrapeAction
	l            logger.Client
	errorHandler func(error)
}

func NewServer(errHandler func(error)) (s *Server, err error) {
	var dbOpts db.Options
	if errHandler == nil {
		errHandler = defaultErrorHandler
	}
	db, err := db.New(dbOpts)
	if err != nil {
		return nil, err
	}
	s = &Server{
		router:       mux.NewRouter(),
		v:            gocache.NewItemCache().WithDb(db),
		db:           db,
		errorHandler: errHandler,
		scrapeQueue:  make(chan scrapeAction, 1<<12),
	}
	return
}

func randomWait() {
	n, _ := rand.Int(bytes.NewReader([]byte{}), big.NewInt(100))
	time.Sleep(time.Duration(n.Int64()) * time.Second)
}

func (s *Server) processQueue() {

	for {
		if err := s.handleScrapeAction(<-s.scrapeQueue); err != nil {
			s.errorHandler(err)
		}
		randomWait()
	}
}

func defaultErrorHandler(err error) {
	fmt.Println(err)
}

func (s *Server) handleScrapeAction(a scrapeAction) error {

	return nil
}
