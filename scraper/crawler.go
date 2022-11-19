package scraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	gocache "github.com/mrod502/go-cache"
	"github.com/mrod502/logger"
	"go.uber.org/atomic"
)

var (
	ptnFindHref = regexp.MustCompile(`href="([^"]+)"`)
)

type ResponseHandler func(*http.Response) error

type CrawlerParams struct {
	Limit uint64
	Seeds []string
}

type Crawler struct {
	l                  logger.Client
	visited            *gocache.Cache[string, bool]
	seeds              *gocache.Cache[string, bool]
	applicationHandler func(*http.Response) error
	textHandler        func(*http.Response) error
	requestCount       *atomic.Uint64
	maxRequests        uint64
	wg                 *sync.WaitGroup
	workers            []chan string
	currentWorker      *atomic.Uint32
	ignore             *gocache.Cache[string, bool]
}

func NewCrawler(params CrawlerParams) (c *Crawler, err error) {
	if len(params.Seeds) == 0 {
		return nil, errors.New("cannot initialize with no seeds")
	}
	c = &Crawler{
		visited:       gocache.New[string, bool](),
		seeds:         gocache.New[string, bool](),
		requestCount:  atomic.NewUint64(0),
		maxRequests:   params.Limit,
		wg:            &sync.WaitGroup{},
		currentWorker: atomic.NewUint32(0),
		ignore: gocache.New(map[string]bool{
			//set some ignored patterns so we don't get blacklisted :)
			"googlesyndication": true,
			"/video":            true,
			"viewkey":           true,
			"youtube.com":       true,
			"adservices":        true,
			"adsense":           true,
			"google":            true,
			".png":              true,
			".ico":              true,
			".js":               true,
			"reddit.com":        true,
		}),
	}
	for _, seed := range params.Seeds {
		c.seeds.Set(seed, true)
	}
	return
}

func (c *Crawler) SetApplicationFileHandler(f ResponseHandler) {
	c.applicationHandler = f
}

func (c *Crawler) SetTextFileHandler(f ResponseHandler) {
	c.textHandler = f
}

func (c *Crawler) Crawl(numWorkers uint8) error {
	if numWorkers < 1 {
		numWorkers = 1
	}
	c.wg.Add(1)
	c.buildWorkers(numWorkers)
	for _, seed := range c.seeds.GetKeys() {
		c.dispatchWork(seed)
	}
	c.wg.Wait()

	return nil
}

func (c *Crawler) dispatchWork(uri string) {
	for {
		if val := c.currentWorker.Load(); cap(c.workers[val]) > len(c.workers[val]) {
			c.workers[val] <- uri
			c.currentWorker.Store((val + 1) % uint32(len(c.workers)))
			return
		} else {
			c.currentWorker.Store((val + 1) % uint32(len(c.workers)))
		}
	}
}

func (c *Crawler) buildWorkers(numWorkers uint8) {
	c.workers = make([]chan string, 0, int(numWorkers))

	for i := 0; i < int(numWorkers); i++ {
		c.workers = append(c.workers, make(chan string, 1<<16))
		go func(ix int) {
			for c.requestCount.Load() < c.maxRequests {
				href := <-c.workers[ix]
				c.findHrefs(href)
			}
		}(i)
	}
}

// used within a worker goroutine
func (c *Crawler) findHrefs(uri string) {
	c.log(uri)
	if v, _ := c.visited.Get(uri); v {
		return
	}

	var cli *http.Client = new(http.Client)
	*cli = *http.DefaultClient
	c.visited.Set(uri, true)

	res, err := cli.Get(uri)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	c.requestCount.Add(1)

	if ct := res.Header.Get("content-type"); strings.Contains(ct, "text/") {

		if c.textHandler != nil {
			c.textHandler(res)
		}
	} else if strings.Contains(ct, "application/") {
		if c.applicationHandler != nil {
			c.applicationHandler(res)
		}
	}
	if c.requestCount.Load() == c.maxRequests {
		c.wg.Done()
		return
	}

	for _, match := range ptnFindHref.FindAllSubmatch(b, -1) {
		ms := string(match[1])
		if len(ms) > 2 && ms[0:2] == "//" {
			ms = "https://" + ms[2:]
		}
		if !c.ignored(ms) {
			if !strings.Contains(ms, "http") {
				ms = strings.TrimSuffix(uri, "/") + "/" + strings.TrimPrefix(ms, "/")
			}
			if strings.Contains(ms, "http") {
				c.dispatchWork(ms)
			}
		}
	}
}

func (c *Crawler) SetLogger(l logger.Client) {
	c.l = l
}

func (c *Crawler) log(l ...string) {
	if c.l != nil {
		c.l.Write(append([]string{"crawler"}, l...)...)
	}
}

func (c *Crawler) ignored(uri string) bool {
	for _, v := range c.ignore.GetKeys() {
		if strings.Contains(uri, v) {
			return true
		}
	}
	return false
}
