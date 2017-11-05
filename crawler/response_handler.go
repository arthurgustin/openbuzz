package crawler

import (
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"github.com/arthurgustin/openbuzz/orm"
	"github.com/arthurgustin/openbuzz/shared"
	"github.com/badoux/checkmail"
	"github.com/m1ome/leven"
	"github.com/mvdan/xurls"
)

type ResponseHandler struct {
	prospect         *orm.Prospect
	fetchbotHandler  fetchbot.HandlerFunc
	mu               sync.Mutex
	alreadyVisited   map[string]bool
	socialStrategies []SocialStrategy
	Logger           shared.LoggerInterface `inject:""`
	Config           *shared.AppConfig      `inject:""`
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
	h.parseHead(ctx, doc)
	h.parseBody(ctx, doc)
}

func (h *ResponseHandler) parseBody(ctx *fetchbot.Context, doc *goquery.Document) {
	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		h.parseHrefBody(ctx, s)
	})
}

func (h *ResponseHandler) parseHrefBody(ctx *fetchbot.Context, s *goquery.Selection) {
	s.Find("a[href]").Each(func(i int, s *goquery.Selection) {
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

		h.fillProspectInformations(url)

		if strings.Contains(h.prospect.GetUrl(), u.Host) {
			if !h.alreadyVisited[url] {
				if _, err := ctx.Q.SendStringGet(url); err != nil {
					h.Logger.Warn(err.Error(), "url", url)
				} else {
					h.Logger.Info("sent string get", "url", url)
					h.mu.Lock()
					h.alreadyVisited[url] = true
					h.mu.Unlock()
				}
			}
		}
	})
}

func (h *ResponseHandler) parseHead(ctx *fetchbot.Context, doc *goquery.Document) {
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		s.Find("link[href]").Each(func(j int, s *goquery.Selection) {
			link, _ := s.Attr("href")

			link = h.decodeURIComponent(link)

			if h.isLinkAnImage(link) {
				h.Logger.Info("FOUND ICON: " + link)
				h.prospect.SetIcon(link)
			}
		})
		s.Find(`meta[name="keywords"]`).Each(func(j int, s *goquery.Selection) {
			val, _ := s.Attr("content")
			val = h.decodeURIComponent(val)

			tags := strings.Split(val, ",")
			for _, tag := range tags {
				tag = strings.Trim(tag, " ")
				h.Logger.Info("FOUND TAG: " + tag)
				h.prospect.SetTag(tag)
			}
		})
		s.Find(`meta[name="description"]`).Each(func(j int, s *goquery.Selection) {
			content, _ := s.Attr("content")
			content = h.decodeURIComponent(content)
			h.Logger.Info("FOUND DESCRIPTION: " + content)

			h.prospect.SetDescription(content)
		})
	})
}

func (h *ResponseHandler) decodeURIComponent(str string) string {
	replacer := strings.NewReplacer("%20", " ", "%21", "!", "%27", "'", "%28", "(", "%29", ")", "%2A", "*")
	return replacer.Replace(str)
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
		if !strings.Contains(targetUrl, socialStrategy.GetUrlPrefix()) {
			continue
		}
		s := strings.Split(targetUrl, socialStrategy.GetUrlPrefix())
		if len(s) > 1 {
			confidence = h.normalizedLevenstein(s[1], h.prospect.GetDomainNameWithoutExtension())
		}
		//if confidence > 0 {
		h.prospect.SetSocial(socialStrategy.GetName(), targetUrl, confidence)
		//}
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
