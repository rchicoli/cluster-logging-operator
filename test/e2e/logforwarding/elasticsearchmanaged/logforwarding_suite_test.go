package elasticsearchmanaged

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterLogForwarder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterLogForwarder E2E Suite - Elasticsearch Managed")
}
