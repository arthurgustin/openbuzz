package orm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"open-buzz/orm"
)

var (
	returnedError error
	client *orm.Client
	err error
)


var _ = Describe("Client", func() {

	BeforeEach(func(){
		client, err = orm.NewClient()
		if err != nil {
			panic(err)
		}
	})

	Context("List", func(){

		var returnedProspects []orm.Prospect

		JustBeforeEach(func(){
			returnedProspects, returnedError = client.List()
		})

		It("should returns something", func(){
			Expect(returnedProspects).To(Equal([]orm.Prospect{

			}))
		})

		It("should not returns an error", func(){
			Expect(returnedError).To(BeNil())
		})
	})

})
