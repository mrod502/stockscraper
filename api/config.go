package api

import (
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
	"github.com/mrod502/stockscraper/obj/document"
	"github.com/mrod502/stockscraper/scraper"
)

type Config struct {
	Obj       document.Config
	Scraper   scraper.Config
	Db        db.Config
	ServePort uint16
	Logger    logger.ClientConfig
}
