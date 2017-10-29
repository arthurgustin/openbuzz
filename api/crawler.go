package api

import (
	"net/http"
	"fmt"
	"encoding/json"
	"open-buzz/crawler"
)

type CrawlerHandler struct {
	Crawler interface {
		CrawlWebsite(url string) (crawler.CrawlResponse, error)
	} `ìnject:""`
}

type requestCrawl struct {
	TargetUrl string `json:"targetUrl"`
}

func writeError(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(400)
	writeJson(w, data)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(200)
	writeJson(w, data)
}

func writeJson(w http.ResponseWriter, data interface{}) {
	toWrite, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	w.Write(toWrite)
}

func (c *CrawlerHandler) CrawlWebsite(w http.ResponseWriter, r *http.Request) {
	fmt.Println("crawl...")
	target := requestCrawl{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&target); err != nil {
		writeError(w, err.Error())
	}

	resp, err := c.Crawler.CrawlWebsite(target.TargetUrl)
	if err != nil {
		writeError(w, err.Error())
	}

	writeSuccess(w, resp)

	return
}

