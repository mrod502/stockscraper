package api

import (
	"net/http"
)

func (s *Server) buildRoutes() {
	s.router.HandleFunc("/scrape/{symbol}/{filetype}", s.withMiddleware(s.Scrape))
	s.router.HandleFunc("/query", s.withMiddleware(s.Query))
	s.router.HandleFunc("/crawl", s.withMiddleware(s.crawl))
}

func (s *Server) withMiddleware(f HttpHandlerInner) HttpHandler {
	return withEnableCors(
		toHttpHandler(
			withErrorHandling(
				s.withLogging(
					f,
				),
			),
		),
	)
}

func (s *Server) withLogging(f HttpHandlerInner) HttpHandlerInner {
	return func(w http.ResponseWriter, r *http.Request) *ResponseError {
		s.log(r.RequestURI, r.RemoteAddr)
		if err := f(w, r); err != nil {
			s.err(r.RequestURI, r.RemoteAddr, err.Error())
			return err
		}
		return nil
	}
}

func withEnableCors(f HttpHandler) HttpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "privatekey")
		w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS,POST,HEAD,DELETE,PUT")
		f(w, r)
	}
}

func withErrorHandling(f HttpHandlerInner) HttpHandlerInner {
	return func(w http.ResponseWriter, r *http.Request) *ResponseError {
		if err := f(w, r); err != nil {
			http.Error(w, err.Message, err.Code)
			return err
		}
		return nil
	}
}

func toHttpHandler(f HttpHandlerInner) HttpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}
