package crawler

import (
	"fmt"
	"github.com/PuerkitoBio/fetchbot"
	"github.com/arthurgustin/openbuzz/shared"
	"net/http"
	"time"
)

func NewFetch(mux *fetchbot.Mux, logger shared.LoggerInterface) *Fetcher {
	h := logHandler(mux, logger)

	fetcher := &Fetcher{
		fetcher:     fetchbot.New(h),
		stopAfter:   time.Duration(15 * time.Second),
		cancelAfter: time.Duration(15 * time.Second),
		memStats:    time.Duration(0 * time.Second),
		Logger:      logger,
	}

	return fetcher
}

// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler, logger shared.LoggerInterface) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			logger.Info("fetch", "code", fmt.Sprintf("%d", res.StatusCode), "method", ctx.Cmd.URL().String(), "content-type", res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

type Fetcher struct {
	fetcher                          *fetchbot.Fetcher
	stopAfter, cancelAfter, memStats time.Duration
	Logger                           shared.LoggerInterface `inject:""`
	Config                           *shared.AppConfig      `inject:""`
}

func (f *Fetcher) Fetch(targetUrl string) {
	queue := f.fetcher.Start()

	// if a stop or cancel is requested after some duration, launch the goroutine
	// that will stop or cancel.
	if f.stopAfter > 0 || f.cancelAfter > 0 {
		after := f.stopAfter
		stopFunc := queue.Close
		if f.cancelAfter != 0 {
			after = f.cancelAfter
			stopFunc = queue.Cancel
		}

		go func() {
			c := time.After(after)
			<-c
			stopFunc()
		}()
	}

	// Enqueue the seed, which is the first entry in the dup map
	_, err := queue.SendStringGet(targetUrl)
	if err != nil {
		f.Logger.Warn(err.Error(), "url", targetUrl)
	}
	queue.Block()

}

// stopHandler stops the fetcher if the stopurl is reached. Otherwise it dispatches
// the call to the wrapped Handler.
func (f *Fetcher) stopHandler(stopurl string, cancel bool, wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if ctx.Cmd.URL().String() == stopurl {
			f.Logger.Info("STOP URL", "url", ctx.Cmd.URL().String())
			// generally not a good idea to stop/block from a handler goroutine
			// so do it in a separate goroutine
			go func() {
				if cancel {
					ctx.Q.Cancel()
				} else {
					ctx.Q.Close()
				}
			}()
			return
		}
		wrapped.Handle(ctx, res, err)
	})
}
