package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("ResourceKindMapper", func() {
	var mapper *ResourceKindMapper

	BeforeEach(func() {
		mapper = &ResourceKindMapper{k8sClient.RESTMapper()}
	})

	It("should map GRs to GVKs", func() {
		kinds, err := mapper.NamespacedKindsForResources("deployments.apps", "jobs.batch", "pods")
		Expect(err).ToNot(HaveOccurred())

		expectedGVKs := []schema.GroupVersionKind{
			{Group: "apps", Version: "v1", Kind: "Deployment"},
			{Group: "batch", Version: "v1", Kind: "Job"},
			{Version: "v1", Kind: "Pod"},
		}
		Expect(kinds).To(Equal(expectedGVKs))
	})

	It("should map NO GRs to NO GVKs", func() {
		kinds, err := mapper.NamespacedKindsForResources()
		Expect(err).ToNot(HaveOccurred())
		Expect(kinds).To(BeEmpty())
	})

	It("should error for invalid resource", func() {
		_, err := mapper.NamespacedKindsForResources("foo")
		Expect(err).To(HaveOccurred())
	})
})
