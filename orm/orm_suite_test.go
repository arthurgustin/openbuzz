package orm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOrm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Orm Suite")
}
