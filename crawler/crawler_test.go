package crawler_test

import (
	. "github.com/arthurgustin/openbuzz/crawler"
	"github.com/arthurgustin/openbuzz/orm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	crawler       *Crawler
	crawlResponse CrawlResponse
	returnedError error
	crawlInfo     CrawlInputInformations
	dbClient      *orm.Client
	err           error
)

var _ = BeforeSuite(func() {
	dbClient, err = orm.NewClient()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if dbClient != nil && dbClient.Db != nil {
		dbClient.Db.Close()
	}
})

var _ = Describe("Crawler", func() {

	var targetUrl string

	BeforeEach(func() {
		crawler = &Crawler{
			DbClient:    dbClient,
			EmailFinder: &EmailFinder{},
		}
	})

	JustBeforeEach(func() {
		crawlInfo = CrawlInputInformations{
			TargetUrl: targetUrl,
		}
		crawlResponse, returnedError = crawler.CrawlWebsite(crawlInfo)
	})

	assertSucces := func() {
		It("should not return an error", func() {
			Expect(returnedError).To(BeNil())
		})
	}

	Context("https://www.kobaltmusic.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.kobaltmusic.com"
		})
		assertSucces()
	})
	Context("https://www.searchenginejournal.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.searchenginejournal.com"
		})
		assertSucces()
	})
	Context("http://www.copyblogger.com", func() {
		BeforeEach(func() {
			targetUrl = "http://www.copyblogger.com"
		})
		assertSucces()
	})
	Context("https://makeawebsitehub.com", func() {
		BeforeEach(func() {
			targetUrl = "https://makeawebsitehub.com"
		})
		assertSucces()
	})
	Context("https://www.impactbnd.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.impactbnd.com"
		})
		assertSucces()
	})
	Context("https://www.tipsandtricks-hq.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.tipsandtricks-hq.com"
		})
		assertSucces()
	})
	Context("https://www.brafton.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.brafton.com"
		})
		assertSucces()
	})
	Context("https://www.quicksprout.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.quicksprout.com"
		})
		assertSucces()
	})
	Context("http://www.bryaneisenberg.com", func() {
		BeforeEach(func() {
			targetUrl = "http://www.bryaneisenberg.com"
		})
		assertSucces()
	})
	Context("http://www.johnchow.com/blog", func() {
		BeforeEach(func() {
			targetUrl = "http://www.johnchow.com/blog"
		})
		assertSucces()
	})
	Context("https://amylynnandrews.com", func() {
		BeforeEach(func() {
			targetUrl = "https://amylynnandrews.com"
		})
		assertSucces()
	})
	Context("https://kaiserthesage.com", func() {
		BeforeEach(func() {
			targetUrl = "https://kaiserthesage.com"
		})
		assertSucces()
	})
	Context("https://problogger.com", func() {
		BeforeEach(func() {
			targetUrl = "https://problogger.com"
		})
		assertSucces()
	})
	Context("https://raventools.com", func() {
		BeforeEach(func() {
			targetUrl = "https://raventools.com"
		})
		assertSucces()
	})
	Context("https://www.postplanner.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.postplanner.com"
		})
		assertSucces()
	})
	Context("https://www.matthewwoodward.co.uk", func() {
		BeforeEach(func() {
			targetUrl = "https://www.matthewwoodward.co.uk"
		})
		assertSucces()
	})
	Context("http://thecopybot.com", func() {
		BeforeEach(func() {
			targetUrl = "http://thecopybot.com"
		})
		assertSucces()
	})
	Context("https://www.thesaleslion.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.thesaleslion.com"
		})
		assertSucces()
	})
	Context("https://smartblogger.com/", func() {
		BeforeEach(func() {
			targetUrl = "https://smartblogger.com/"
		})
		assertSucces()
	})
	Context("https://www.analyticsvidhya.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.analyticsvidhya.com"
		})
		assertSucces()
	})
	Context("http://online-behavior.com", func() {
		BeforeEach(func() {
			targetUrl = "http://online-behavior.com"
		})
		assertSucces()
	})
	Context("https://marketingland.com", func() {
		BeforeEach(func() {
			targetUrl = "https://marketingland.com"
		})
		assertSucces()
	})
	Context("https://www.shoutmeloud.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.shoutmeloud.com"
		})
		assertSucces()
	})
	Context("https://yourstory.com", func() {
		BeforeEach(func() {
			targetUrl = "https://yourstory.com"
		})
		assertSucces()
	})
	Context("https://neilpatel.com", func() {
		BeforeEach(func() {
			targetUrl = "https://neilpatel.com"
		})
		assertSucces()
	})
	Context("https://backlinko.com", func() {
		BeforeEach(func() {
			targetUrl = "https://backlinko.com"
		})
		assertSucces()
	})
	Context("https://ahrefs.com", func() {
		BeforeEach(func() {
			targetUrl = "https://ahrefs.com"
		})
		assertSucces()
	})
	Context("https://www.kobaltmusic.com", func() {
		BeforeEach(func() {
			targetUrl = "https://www.kobaltmusic.com"
		})
		assertSucces()
	})

})
