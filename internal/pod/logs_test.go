package pod

import (
	"context"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("LogsReader", func() {
	var (
		ctx    context.Context
		reader LogsReader
	)

	BeforeEach(func() {
		ctx = context.Background()
		reader = NewLogsReader(fake.NewSimpleClientset())
	})

	Context("GetLogs", func() {
		It("should return logs from pod container", func() {
			logs, err := reader.GetLogs(ctx, types.NamespacedName{}, "container")
			Expect(err).NotTo(HaveOccurred())
			Expect(io.ReadAll(logs)).To(Equal([]byte("fake logs")))
		})
	})
})
