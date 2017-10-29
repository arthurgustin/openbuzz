package crawler

import (
	"fmt"
	"net/url"
	"net/http"

	"open-buzz/orm"
	"github.com/PuerkitoBio/fetchbot"
	"open-buzz/shared"
)

type Crawler struct {
	DbClient *orm.Client
	*EmailFinder `inject:""`
}

type CrawlResponse struct {
	SocialNetworks struct {
		Google []string `json:"google"`
		Youtube []string `json:"youtube"`
		Twitter []string `json:"twitter"`
		Facebook []string `json:"facebook"`
	} `json:"socialNetworks"`
	Email []string `json:"email"`
}

type CrawlInputInformations struct {
	TargetUrl, FirstName, MiddleName, LastName string
}

func (c *Crawler) CrawlWebsite(input CrawlInputInformations) (CrawlResponse, error) {
	prospect := orm.NewProspect(input.TargetUrl).SetFirstName(input.FirstName).SetMiddleName(input.MiddleName).SetLastName(input.LastName)

	responseHandler := &ResponseHandler{
		prospect: prospect,
		alreadyVisited: map[string]bool{
			input.TargetUrl: true,
		},
		socialStrategies: GetAllSocialStrategies(),
		Logger: shared.NewLogger(),
	}

	mux := c.NewMux(responseHandler)

	f := NewFetch(mux)
	f.Fetch(input.TargetUrl)

	emails, err := c.EmailFinder.Find(*prospect)
	if err != nil {
		switch err {
		case ErrAllPolicyActivated:
			fmt.Println(ErrAllPolicyActivated)
		}
	}
	for _, email := range emails {
		prospect.SetEmail(email.email, 1.)
	}

	if err = c.DbClient.Save(prospect); err != nil {
		return CrawlResponse{}, err
	}

	return CrawlResponse{}, nil
}

func (c *Crawler) NewMux(responseHandler *ResponseHandler) *fetchbot.Mux {
	// Create the muxer
	mux := fetchbot.NewMux()

	// Handle all errors the same
	mux.HandleErrors(fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		fmt.Printf("[ERR] %s %s - %s\n", ctx.Cmd.Method(), ctx.Cmd.URL(), err)
	}))

	// Handle GET requests for html responses, to parse the body and enqueue all links as HEAD
	// requests.
	mux.Response().Method("GET").ContentType("text/html").Handler(responseHandler.getHandler())

	// Handle HEAD requests for html responses coming from the source host - we don't want
	// to crawl links from other hosts.
	u, err := url.Parse(responseHandler.prospect.GetUrl())
	if err != nil {
		panic(err)
	}
	mux.Response().Method("HEAD").Host(u.Host).ContentType("text/html").Handler(responseHandler.headHandler())

	return mux
}