package pod

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

//go:generate go run -tags generate github.com/vektra/mockery/v2 --name LogsReader --filename zz_generated.pod_mocks.go --inpackage --with-expecter
type LogsReader interface {
	GetLogs(ctx context.Context, pod types.NamespacedName, container string) (io.ReadCloser, error)
}

func NewLogsReader(clientset kubernetes.Interface) *logsReader {
	return &logsReader{clientset: clientset}
}

type logsReader struct {
	clientset kubernetes.Interface
}

func (l logsReader) GetLogs(ctx context.Context, pod types.NamespacedName, container string) (io.ReadCloser, error) {
	return l.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: container,
	}).Stream(ctx)
}

var _ LogsReader = logsReader{}
