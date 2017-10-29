package crawler

import (
	"github.com/PuerkitoBio/fetchbot"
	"runtime"
	"bytes"
	"strings"
	"fmt"
	"time"
	"sync"
	"net/http"
)

func NewFetch(mux *fetchbot.Mux) *Fetcher {
	h := logHandler(mux)

	fetcher := &Fetcher{
		fetcher: fetchbot.New(h),
		stopAfter: time.Duration(30*time.Second),
		cancelAfter: time.Duration(30*time.Second),
		memStats: time.Duration(0*time.Second),
	}

	// First mem stat print must be right after creating the fetchbot
	if fetcher.memStats > 0 {
		// Print starting stats
		fetcher.printMemStats(nil)
		// Run at regular intervals
		fetcher.runMemStats()
		// On exit, print ending stats after a GC
		defer func() {
			runtime.GC()
			fetcher.printMemStats(nil)
		}()
	}

	return fetcher
}


// logHandler prints the fetch information and dispatches the call to the wrapped Handler.
func logHandler(wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if err == nil {
			fmt.Printf("[%d] %s %s - %s\n", res.StatusCode, ctx.Cmd.Method(), ctx.Cmd.URL(), res.Header.Get("Content-Type"))
		}
		wrapped.Handle(ctx, res, err)
	})
}

type Fetcher struct {
	fetcher *fetchbot.Fetcher
	stopAfter, cancelAfter, memStats time.Duration
}

func (f *Fetcher) runMemStats() {
	var mu sync.Mutex
	var di *fetchbot.DebugInfo

	// Start goroutine to collect fetchbot debug info
	go func() {
		for v := range f.fetcher.Debug() {
			mu.Lock()
			di = v
			mu.Unlock()
		}
	}()
	// Start ticker goroutine to print mem stats at regular intervals
	go func() {
		for _ = range time.Tick(f.memStats) {
			mu.Lock()
			f.printMemStats(di)
			mu.Unlock()
		}
	}()
}

func (f *Fetcher)  printMemStats(di *fetchbot.DebugInfo) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	buf := bytes.NewBuffer(nil)
	buf.WriteString(strings.Repeat("=", 72) + "\n")
	buf.WriteString("Memory Profile:\n")
	buf.WriteString(fmt.Sprintf("\tAlloc: %d Kb\n", mem.Alloc/1024))
	buf.WriteString(fmt.Sprintf("\tTotalAlloc: %d Kb\n", mem.TotalAlloc/1024))
	buf.WriteString(fmt.Sprintf("\tNumGC: %d\n", mem.NumGC))
	buf.WriteString(fmt.Sprintf("\tGoroutines: %d\n", runtime.NumGoroutine()))
	if di != nil {
		buf.WriteString(fmt.Sprintf("\tNumHosts: %d\n", di.NumHosts))
	}
	buf.WriteString(strings.Repeat("=", 72))
	fmt.Println(buf.String())
}

// stopHandler stops the fetcher if the stopurl is reached. Otherwise it dispatches
// the call to the wrapped Handler.
func (f *Fetcher)  stopHandler(stopurl string, cancel bool, wrapped fetchbot.Handler) fetchbot.Handler {
	return fetchbot.HandlerFunc(func(ctx *fetchbot.Context, res *http.Response, err error) {
		if ctx.Cmd.URL().String() == stopurl {
			fmt.Printf(">>>>> STOP URL %s\n", ctx.Cmd.URL())
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

func (f *Fetcher) Fetch(targetUrl string) {
	fmt.Println("about to fetch " + targetUrl)
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
		fmt.Printf("[ERR] GET %s - %s\n", targetUrl, err)
	}
	queue.Block()

}