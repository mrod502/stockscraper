package api

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

func randomWait() {
	n, _ := rand.Int(bytes.NewReader([]byte{}), big.NewInt(100))
	time.Sleep(time.Duration(n.Int64()) * time.Second)
}

func requestSummary(r *http.Request) []string {
	return []string{
		r.RemoteAddr,
		r.Proto,
		r.Method,
		r.URL.EscapedPath(),
		strconv.Itoa(int(r.ContentLength)),
	}
}

func (s *Server) err(v ...string) {
	s.l.Write(append([]string{"SCRAPER", "API", "ERR"}, v...)...)
}

func (s *Server) log(v ...string) {
	s.l.Write(append([]string{"SCRAPER", "API", "LOG"}, v...)...)
}
