package crawler_test

import (
	. "open-buzz/crawler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"open-buzz/orm"
)

var (
	crawler *Crawler
	targetUrl string
	crawlResponse CrawlResponse
	returnedError error
	crawlInfo CrawlInputInformations
)


var _ = Describe("Crawler", func() {
	BeforeEach(func(){
		dbClient, err := orm.NewClient()
		if err != nil {
			panic(err)
		}
		crawler = &Crawler{
			DbClient: dbClient,
			EmailFinder: &EmailFinder{},
		}
	})

	JustBeforeEach(func(){
		crawlResponse, returnedError = crawler.CrawlWebsite(crawlInfo)
	})

	FContext("kobaltmusic", func(){
		BeforeEach(func(){
			crawlInfo = CrawlInputInformations{
				TargetUrl: "https://www.kobaltmusic.com",
			}
		})

		It("should not return an error", func(){
			Expect(returnedError).To(BeNil())
		})

	})

	Context("ahrefs", func(){
		BeforeEach(func(){
			crawlInfo = CrawlInputInformations{
				TargetUrl: "https://ahrefs.com",
			}
		})

		It("should not return an error", func(){
			Expect(returnedError).To(BeNil())
		})

	})

	Context("backlinko", func(){
		BeforeEach(func(){
			crawlInfo = CrawlInputInformations{
				TargetUrl: "https://backlinko.com",
			}
		})

		It("should not return an error", func(){
			Expect(returnedError).To(BeNil())
		})

	})

	Context("neilpatel.com", func(){
		BeforeEach(func(){
			crawlInfo = CrawlInputInformations{
				TargetUrl: "https://neilpatel.com",
			}
		})

		It("should not return an error", func(){
			Expect(returnedError).To(BeNil())
		})

	})

	Context("yourstory", func(){
		BeforeEach(func(){
			crawlInfo = CrawlInputInformations{
				TargetUrl: "https://yourstory.com/",
			}
		})

		It("should not return an error", func(){
			Expect(returnedError).To(BeNil())
		})

	})

})
