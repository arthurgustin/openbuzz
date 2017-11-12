package main

import (
	"fmt"
	"github.com/allan-simon/go-singleinstance"
	"github.com/arthurgustin/openbuzz/api"
	"github.com/arthurgustin/openbuzz/crawler"
	"github.com/arthurgustin/openbuzz/orm"
	"github.com/arthurgustin/openbuzz/shared"
	"github.com/facebookgo/inject"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/cors"
	"net/http"
)

var (
	logger    shared.LoggerInterface
	appConfig = &shared.AppConfig{}
	dbClient  = &orm.Client{}
)

const configPrefix = "OPENBUZZ"

func main() {
	logger = shared.NewLogger()

	if err := envconfig.Process(configPrefix, appConfig); err != nil {
		logger.Fatal(err.Error())
	}

	_, err := singleinstance.CreateLockFile("buzz.lock")
	if err != nil {
		logger.Fatal("an instance already exists")
		return
	}

	crawlerHandler := &api.CrawlerHandler{}
	webCrawler := &crawler.Crawler{}
	prospectorHandler := &api.ProspectHandler{}
	if err := inject.Populate(appConfig, crawlerHandler, webCrawler, dbClient, logger, prospectorHandler); err != nil {
		logger.Fatal(err.Error())
		return
	}

	if err := dbClient.Init(); err != nil {
		logger.Fatal(err.Error())
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/crawl", crawlerHandler.CrawlWebsite).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/list", prospectorHandler.List).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/prospect/{prospectId}", prospectorHandler.Delete).Methods(http.MethodDelete)
	handler := cors.AllowAll().Handler(r)

	logger.Info("starting listening...", "port", fmt.Sprintf("%d", appConfig.Port))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", appConfig.Port), handler); err != nil {
		logger.Fatal("ListenAndServe: ", "err", err.Error())
	}
}
