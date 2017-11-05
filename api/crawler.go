package api

import (
	"encoding/json"
	"fmt"
	"github.com/arthurgustin/openbuzz/crawler"
	"github.com/arthurgustin/openbuzz/shared"
	"net/http"
	"sync"
	"time"
)

type CrawlerHandler struct {
	Crawler interface {
		CrawlWebsite(input crawler.CrawlInputInformations) (crawler.CrawlResponse, error)
	} `inject:""`
	Logger shared.LoggerInterface `inject:""`
}

type requestCrawl struct {
	TargetUrls []string `json:"targetUrls"`
}

func writeError(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(400)
	writeJson(w, data)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(200)
	writeJson(w, data)
}

func writeAccepted(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusAccepted)
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
	start := time.Now()
	c.Logger.Info("new incoming crawling request")
	target := requestCrawl{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&target); err != nil {
		writeError(w, err.Error())
	}

	if len(target.TargetUrls) < 1 {
		c.Logger.Info("no urls provided")
		writeError(w, "no urls provided")
		return
	}
	c.Logger.Info(fmt.Sprintf("I started crawling %d websites, come back in a couple of minutes", len(target.TargetUrls)))

	resp := c._crawl(w, start, target)

	writeSuccess(w, resp)

	return
}

type apiCrawlResponse struct {
	NumberOfSuccess int64         `json:"numberOfSuccess"`
	NumberOfFails   int64         `json:"numberOfFails"`
	Details         []crawlDetail `json:"details"`
}

type crawlDetail struct {
	Url    string `json:"url"`
	Reason string `json:"reason"`
	Error  bool   `json:"error"`
}

func (c *CrawlerHandler) _crawl(w http.ResponseWriter, start time.Time, target requestCrawl) (resp apiCrawlResponse) {
	var wg sync.WaitGroup
	wg.Add(len(target.TargetUrls))

	details := make([]crawlDetail, len(target.TargetUrls))

	for i, targetUrl := range target.TargetUrls {
		go func(url string) {
			defer wg.Done()

			c.Logger.Info(fmt.Sprintf("crawling %s", url))
			_, err := c.Crawler.CrawlWebsite(crawler.CrawlInputInformations{
				TargetUrl: url,
			})
			if err != nil {
				c.Logger.Warn(err.Error())
				details[i] = crawlDetail{
					Error:  true,
					Url:    url,
					Reason: err.Error(),
				}
			} else {
				details[i] = crawlDetail{
					Error: false,
					Url:   url,
				}
			}
			c.Logger.Info(fmt.Sprintf("%s has been crawled", url))
		}(targetUrl)
	}
	wg.Wait()
	t := time.Now()
	elapsed := t.Sub(start)
	c.Logger.Info(fmt.Sprintf("DONE: %d websites. Elapsed: %s", len(target.TargetUrls), elapsed.String()))

	resp.Details = details
	for _, d := range details {
		if d.Error {
			resp.NumberOfFails += 1
		} else {
			resp.NumberOfSuccess += 1
		}
	}

	return resp
}
