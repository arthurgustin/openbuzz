package main

import (
	"open-buzz/orm"
	"open-buzz/api"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"github.com/facebookgo/inject"
	"open-buzz/shared"
)

var (
	dbClient *orm.Client
	logger shared.LoggerInterface
	err error
)

func main() {
	dbClient, err = orm.NewClient()
	if err != nil {
		panic(err)
	}

	logger = shared.NewLogger()

	crawler := &api.CrawlerHandler{}
	prospector := &api.ProspectHandler{}

	if err := inject.Populate(crawler, dbClient, logger, prospector); err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/crawl", crawler.CrawlWebsite).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/list", prospector.List).Methods(http.MethodGet)

	if err := http.ListenAndServe(":1344", r); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}