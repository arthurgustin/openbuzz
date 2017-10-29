package api_test

import (
	. "open-buzz/api"

	. "github.com/onsi/ginkgo"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/jarcoal/httpmock"
	"open-buzz/orm"
)

var (
	handler *ProspectHandler
)

var _ = BeforeSuite(func() {
	// block all HTTP requests
	httpmock.Activate()
})

var _ = BeforeEach(func() {
	// remove any mocks
	httpmock.Reset()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

var _ = Describe("Prospect", func() {

	BeforeEach(func() {
		dbClient, err := orm.NewClient()
		if err != nil {
			panic(err)
		}
		handler = &ProspectHandler{
			Client: dbClient,
		}

		r := mux.NewRouter()
		r.HandleFunc("/api/v1/prospects/list", handler.List).Methods(http.MethodGet)
	})

	Context("List", func(){
		JustBeforeEach(func(){
		})
	})

})
