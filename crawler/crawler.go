package crawler

import (
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/fetchbot"
	"github.com/arthurgustin/openbuzz/orm"
	"github.com/arthurgustin/openbuzz/shared"
	"github.com/pkg/errors"
)

var (
	ErrTargetUrlEmpty = errors.New("targetUrl cannot be empty")
)

type Crawler struct {
	DbClient    *orm.Client            `inject:""`
	EmailFinder *EmailFinder           `inject:""`
	Logger      shared.LoggerInterface `inject:""`
	Fetcher     *Fetcher               `inject:""`
	Config      *shared.AppConfig      `inject:""`
}

type CrawlResponse struct {
	SocialNetworks struct {
		Google   []string `json:"google"`
		Youtube  []string `json:"youtube"`
		Twitter  []string `json:"twitter"`
		Facebook []string `json:"facebook"`
	} `json:"socialNetworks"`
	Email []string `json:"email"`
}

type CrawlInputInformations struct {
	TargetUrl, FirstName, MiddleName, LastName string
}

func (c *Crawler) CrawlWebsite(input CrawlInputInformations) (CrawlResponse, error) {
	if input.TargetUrl == "" {
		return CrawlResponse{}, ErrTargetUrlEmpty
	}

	prospect := orm.NewProspect(input.TargetUrl).SetFirstName(input.FirstName).SetMiddleName(input.MiddleName).SetLastName(input.LastName)

	responseHandler := &ResponseHandler{
		prospect: prospect,
		alreadyVisited: map[string]bool{
			input.TargetUrl: true,
		},
		socialStrategies: GetAllSocialStrategies(),
		Logger:           c.Logger,
	}

	mux := c.NewMux(responseHandler)

	f := NewFetch(mux, c.Logger)
	f.Fetch(input.TargetUrl)

	emails, err := c.EmailFinder.Find(*prospect)
	if err != nil {
		switch err {
		case ErrAllPolicyActivated:
			c.Logger.Warn(ErrAllPolicyActivated.Error())
		}
	}
	for _, email := range emails {
		prospect.SetEmail(email.email, 0.5)
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
		c.Logger.Warn(err.Error(), "method", ctx.Cmd.Method(), "url", ctx.Cmd.URL().String())
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
