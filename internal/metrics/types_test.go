package metrics

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type fakeManager struct{}

func (f fakeManager) SetFields(i interface{}) error {
	return nil
}

func (f fakeManager) GetConfig() *rest.Config {
	return nil
}

func (f fakeManager) GetScheme() *runtime.Scheme {
	return nil
}

func (f fakeManager) GetClient() client.Client {
	return nil
}

func (f fakeManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (f fakeManager) GetCache() cache.Cache {
	return nil
}

func (f fakeManager) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

func (f fakeManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (f fakeManager) GetAPIReader() client.Reader {
	return nil
}

func (f fakeManager) Start(ctx context.Context) error {
	return nil
}

func (f fakeManager) Add(runnable manager.Runnable) error {
	return nil
}

func (f fakeManager) Elected() <-chan struct{} {
	return nil
}

func (f fakeManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	return nil
}

func (f fakeManager) AddHealthzCheck(name string, check healthz.Checker) error {
	return nil
}

func (f fakeManager) AddReadyzCheck(name string, check healthz.Checker) error {
	return nil
}

func (f fakeManager) GetWebhookServer() *webhook.Server {
	return nil
}

func (f fakeManager) GetLogger() logr.Logger {
	return logr.Logger{}
}

func (f fakeManager) GetControllerOptions() v1alpha1.ControllerConfigurationSpec {
	return v1alpha1.ControllerConfigurationSpec{}
}
