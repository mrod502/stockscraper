package api

import (
	"github.com/mrod502/logger"
	"github.com/mrod502/stockscraper/db"
	"github.com/mrod502/stockscraper/obj"
	"github.com/mrod502/stockscraper/scraper"
)

type Config struct {
	Obj       obj.Config
	Scraper   scraper.Config
	Db        db.Config
	ServePort uint16
	Logger    logger.ClientConfig
}
