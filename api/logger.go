package api

import (
	"fmt"
	"time"

	"github.com/mrod502/logger"
)

type Logger struct {
	queue chan []string
	f     *logger.FileLog
}

func NewLogger(path string) (*Logger, error) {
	l := &Logger{
		queue: make(chan []string, 2048),
	}
	var err error
	l.f, err = logger.NewFileLog(path, make(chan []string, 512))
	go l.f.Start()
	go l.listen()
	return l, err
}

func (l *Logger) listen() {
	for {
		var out = <-l.queue
		l.f.Write(out...)
		fmt.Println(append([]string{time.Now().Format(time.RFC3339)}, out...))
	}
}

func (l Logger) SetLogLocally(bool) {

}
func (l Logger) LogLocally() bool {
	return true
}
func (l Logger) Write(v ...string) error {
	l.queue <- v
	return nil
}
func (l Logger) Connect() error {
	return nil
}
func (l *Logger) Stop() {
	l.f.Stop()
}
