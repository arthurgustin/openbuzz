package crawler

import (
	"open-buzz/orm"
	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"net/http"
	"strings"
	"github.com/badoux/checkmail"
	"sync"
	"github.com/m1ome/leven"
	"github.com/mvdan/xurls"
	"open-buzz/shared"
)

type ResponseHandler struct {
	prospect *orm.Prospect
	fetchbotHandler fetchbot.HandlerFunc
	mu sync.Mutex
	alreadyVisited map[string]bool
	socialStrategies []SocialStrategy
	Logger shared.LoggerInterface `inject:""`
}

func (h *ResponseHandler) headHandler() fetchbot.HandlerFunc {
	return func(ctx *fetchbot.Context, res *http.Response, err error) {
		if _, err := ctx.Q.SendStringGet(ctx.Cmd.URL().String()); err != nil {
			h.Logger.Warn(err.Error(), "method", ctx.Cmd.Method(), "url", ctx.Cmd.URL().String())
		}
	}
}

func (h *ResponseHandler) getHandler() fetchbot.HandlerFunc {
	return func(ctx *fetchbot.Context, res *http.Response, err error) {
		// Process the body to find the links
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			h.Logger.Warn(err.Error(), "method", ctx.Cmd.Method(), "url", ctx.Cmd.URL().String())
			return
		}
		// Enqueue all links as GET requests
		h.enqueueLinks(ctx, doc)
	}
}

func (h *ResponseHandler) enqueueLinks(ctx *fetchbot.Context, doc *goquery.Document) {
	h.mu.Lock()
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("href")
		u, err := ctx.Cmd.URL().Parse(val)
		if err != nil {
			h.Logger.Warn(err.Error(), "method", ctx.Cmd.Method(), "url", val)
			return
		}

		url := xurls.Relaxed.FindString(u.String())
		if url == "" {
			return
		}
		host := u.Host

		h.fillProspectInformations(url)

		if strings.HasPrefix(h.prospect.GetUrl(), host) {
			fmt.Println(u.RawQuery)
			if !h.alreadyVisited[url] {
				if _, err := ctx.Q.SendStringGet(url); err != nil {
					h.Logger.Warn(err.Error(), "url", url)
				} else {
					h.alreadyVisited[url] = true
				}
			}
		}
	})
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		doc.Find("link[href]").Each(func (j int, s *goquery.Selection){
			val, _ := s.Attr("href")
			u, err := ctx.Cmd.URL().Parse(val)
			if err != nil {
				h.Logger.Warn(err.Error())
				return
			}
			link := u.String()

			if h.isLinkAnImage(link) {
				h.Logger.Info(link)
				h.prospect.SetIcon(link)
			}
		})
		doc.Find(`meta[name="keywords"]`).Each(func (j int, s *goquery.Selection){
			val, _ := s.Attr("content")
			u, err := ctx.Cmd.URL().Parse(val)
			if err != nil {
				h.Logger.Warn(err.Error())
				return
			}
			content := u.String()
			tags := strings.Split(content, ",")
			for _, t := range tags {
				h.Logger.Info(strings.Trim(t, " "))
			}

		})
	})
	h.mu.Unlock()
}

func (h *ResponseHandler) isLinkAnImage(link string) bool {
	imgExtensions := []string{".png", ".jpg", ".jpeg", ".bmp", ".ico"}
	for _, ext := range imgExtensions {
		if strings.HasSuffix(link, ext) {
			return true
		}
	}
	return false
}

func (h *ResponseHandler) normalizedLevenstein(a, b string) float64 {
	d1 := leven.Distance(strings.ToLower(a), strings.ToLower(b))
	lenA := len(a)
	lenB := len(b)
	lenMax := float64(lenA)
	if lenB > lenA {
		lenMax = float64(lenB)
	}
	return 1. - (float64(d1) / float64(lenMax))
}

func (h *ResponseHandler) fillProspectInformations(targetUrl string) {
	if strings.Contains(targetUrl, "share") {
		return
	}

	for _, socialStrategy := range h.socialStrategies {
		confidence := 0.
		s := strings.Split(targetUrl, socialStrategy.GetUrlPrefix())
		if len(s) > 1 {
			confidence = h.normalizedLevenstein(s[1], h.prospect.GetHost())
		}
		if confidence > 0.1 {
			h.prospect.SetSocial(socialStrategy.GetName(), targetUrl, confidence)
		}
	}

	if strings.Contains(targetUrl, "@") {
		if err := checkmail.ValidateFormat(targetUrl); err == nil {
			err := checkmail.ValidateHost(targetUrl)
			if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil {
				h.Logger.Warn(smtpErr.Error(), "code", smtpErr.Code())
			} else {
				h.prospect.SetEmail(targetUrl, 1)
			}
		}
	}
}
