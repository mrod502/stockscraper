package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/http"
	"time"

	"github.com/gorilla/mux"
	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
	"github.com/mrod502/stockscraper/obj"
	"github.com/mrod502/stockscraper/scraper"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/atomic"
)

type HttpHandlerInner func(http.ResponseWriter, *http.Request) *ResponseError
type HttpHandler func(http.ResponseWriter, *http.Request)

var (
	ErrLimitReached = errors.New("limit reached")
)

type Server struct {
	router      *mux.Router
	db          *db.DB
	v           *gocache.Cache[string, db.TypedObject]
	l           *Logger
	s           scraper.Client
	c           Config
	newDocsChan chan *obj.Document
	crawlReset  *atomic.Int64
}

func NewServer(cfg Config, errHandler func(error)) (s *Server, err error) {
	dbase, err := db.New(cfg.Db)
	if err != nil {
		return nil, err
	}
	l, err := NewLogger(cfg.Logger.RemoteIP)
	if err != nil {
		return nil, err
	}
	s = &Server{
		router:      mux.NewRouter(),
		v:           gocache.New[string, db.TypedObject](),
		db:          dbase,
		newDocsChan: make(chan *obj.Document, 512),
		l:           l,
		c:           cfg,
		s:           scraper.NewGoogleClient(),
		crawlReset:  atomic.NewInt64(0),
	}
	s.l.SetLogLocally(true)
	for i := 0; i < 5; err = s.l.Connect() {
		if err == nil {
			break
		}
		logger.Error("SCRAPER", "logger", "connect", err.Error(), "retrying...")
		time.Sleep(2 * time.Second)
		i++
	}
	if err != nil {
		return nil, err
	}
	s.buildRoutes()
	return
}
func (s *Server) Close() error {
	s.l.Stop()
	return s.db.Close()
}

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
	s.log("listening on ", fmt.Sprintf(":%d", s.c.ServePort))
	return http.ListenAndServe(fmt.Sprintf(":%d", s.c.ServePort), s.router)
}

func (s *Server) Query(w http.ResponseWriter, r *http.Request) *ResponseError {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return NewResponseError(http.StatusBadRequest, err.Error())
	}

	var q db.DocQuery
	err = json.Unmarshal(b, &q)
	if err != nil {
		return NewResponseError(http.StatusBadRequest, err.Error())
	}

	res, err := s.docQuery(q)

	if err != nil {
		return NewResponseError(http.StatusInternalServerError, err.Error())
	}
	b, _ = json.Marshal(res)

	_, err = w.Write(b)
	if err != nil {
		s.err("query", r.RemoteAddr, r.URL.EscapedPath(), err.Error())
	}
	return nil
}

func (s *Server) docQuery(q db.DocQuery) ([]obj.Document, error) {
	var docs = make([]obj.Document, 0, q.Limit/2)
	var matches uint
	var d = new(obj.Document)

	err := s.db.Each(func(i db.TypedObject) error {
		err := msgpack.Unmarshal(i.Data, d)
		if err != nil {
			return err
		}
		if q.Match(d) {
			docs = append(docs, *d)
			matches++
		}
		if matches == q.Limit {
			return ErrLimitReached
		}
		return nil
	}, "", obj.TypeDocument)
	if err != nil && err != ErrLimitReached {
		return docs, err
	}
	return docs, nil
}

func (s *Server) Scrape(w http.ResponseWriter, r *http.Request) *ResponseError {
	vars := mux.Vars(r)

	symbol := vars["symbol"]
	ftype := vars["filetype"]
	if !(ftype == "pdf" || ftype == "txt" || ftype == "html" || ftype == "xml") {
		return NewResponseError(http.StatusBadRequest, "invalid filetype")
	}
	d, err := s.s.Scrape(symbol, ftype)

	if err != nil {
		return NewResponseError(http.StatusInternalServerError, err.Error())
	}
	for _, v := range d {
		s.newDocsChan <- v
	}
	b, err := json.Marshal(d)
	if err != nil {
		return NewResponseError(http.StatusInternalServerError, err.Error())
	}
	if err != nil {
		return NewResponseError(http.StatusInternalServerError, err.Error())
	}
	w.Write(b)
	return nil

}

func (s *Server) crawl(w http.ResponseWriter, r *http.Request) *ResponseError {
	if s.crawlReset.Load() > time.Now().Unix() {
		w.Header().Set("x-ratelimit-reset", fmt.Sprintf("%d", s.crawlReset.Load()))
		return NewResponseError(http.StatusTooManyRequests, "too many requests")
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return NewResponseError(http.StatusBadRequest, err.Error())
	}

	s.crawlReset.Store(time.Now().Unix() + 60)

	var params scraper.CrawlerParams
	err = json.Unmarshal(b, &params)
	if err != nil {
		return NewResponseError(http.StatusBadRequest, err.Error())
	}

	c, err := scraper.NewCrawler(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return NewResponseError(http.StatusInternalServerError, err.Error())
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
	c.SetLogger(s.l)
	err = c.Crawl(8)
	if err != nil {
		return NewResponseError(http.StatusInternalServerError, "craw failed")
	}
	return nil
}
