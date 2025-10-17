package metrics

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	openreportsv1alpha1 "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/meta"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	k8smetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/config/feature"
)

const (
	Namespace = "image_scanner"
	Subsystem = "container_image"

	// LoggerName is the name used for the metrics collector logger.
	LoggerName = "metrics-collector"
)

var (
	cisResourceLabels = map[string]cisMetricsLabelFunc{
		"namespace": func(cis stasv1alpha1.ContainerImageScan) string {
			return cis.Namespace
		},
		"name": func(cis stasv1alpha1.ContainerImageScan) string {
			return cis.Name
		},
		"image_name": func(cis stasv1alpha1.ContainerImageScan) string {
			return cis.Spec.Name
		},
		"image_digest": func(cis stasv1alpha1.ContainerImageScan) string {
			return cis.Spec.Digest.String()
		},
		"image_tag": func(cis stasv1alpha1.ContainerImageScan) string {
			return cis.Spec.Tag
		},
	}
)

type ImageMetricsCollector struct {
	client.Reader
	config.Config

	Log logr.Logger

	cisLabels       cisLabels
	successDesc     *prometheus.Desc
	issuesDesc      *prometheus.Desc
	patchStatusDesc *prometheus.Desc
}

// Manager is an interface defined for functions actually used in manager.Manager to make it easier to mock.
type Manager interface {
	Add(manager.Runnable) error
}

func (c *ImageMetricsCollector) SetupWithManager(mgr Manager) error {
	if c.Log.GetSink() == nil {
		c.Log = logr.Discard()
	}

	labels := make(cisLabels, 0, len(c.MetricsLabels)+len(cisResourceLabels)+1)

	if len(c.MetricsLabels) > 0 {
		re := regexp.MustCompile("[^a-zA-Z0-9_]+")

		for _, l := range c.MetricsLabels {
			labelKey := l
			cl := cisLabel{
				name: re.ReplaceAllString(labelKey, "_"),
				value: func(cis stasv1alpha1.ContainerImageScan) string {
					return cis.Labels[labelKey]
				},
			}
			labels = append(labels, cl)
		}
	}

	for k, v := range cisResourceLabels {
		labels = append(labels, cisLabel{name: k, value: v})
	}

	c.cisLabels = labels
	c.successDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "scan_success"),
		"Displays whether or not container image scan was a success",
		labels.names(),
		nil,
	)
	c.issuesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "issues"),
		"Number of container image scan issues",
		labels.names("severity"),
		nil,
	)
	c.patchStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "patch_status"),
		"Number of detected container image vulnerabilities grouped by fixed/unfixed",
		labels.names("condition"),
		nil,
	)

	return mgr.Add(c)
}

func (c ImageMetricsCollector) Start(ctx context.Context) error {
	if err := k8smetrics.Registry.Register(c); err != nil {
		return err
	}

	// Block until the context is done.
	<-ctx.Done()
	k8smetrics.Registry.Unregister(c)

	return nil
}

func (c ImageMetricsCollector) NeedLeaderElection() bool {
	return true
}

func (c ImageMetricsCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.successDesc

	descs <- c.issuesDesc

	descs <- c.patchStatusDesc
}

func (c ImageMetricsCollector) Collect(metrics chan<- prometheus.Metric) {
	ctx := context.Background()

	cisList := &stasv1alpha1.ContainerImageScanList{}
	if err := c.List(ctx, cisList, client.InNamespace("")); err != nil {
		c.Log.Error(err, "Failed to list all ContainerImageScans")
		return
	}

	cisLabelValues := make([]string, len(c.cisLabels))
	issuesLabelValues := make([]string, len(cisLabelValues)+1)
	patchStatusLabelValues := make([]string, len(cisLabelValues)+1)

	for _, cis := range cisList.Items {
		for i, l := range c.cisLabels {
			cisLabelValues[i] = l.value(cis)
		}

		copy(issuesLabelValues, cisLabelValues)
		copy(patchStatusLabelValues, cisLabelValues)

		// TODO: We actually have 3 states here: NotScanned, Reconciling, Stalled; How to represent this in metrics?
		successValue := float64(1)
		if meta.IsStatusConditionTrue(cis.Status.Conditions, string(kstatus.ConditionStalled)) {
			successValue = float64(0)
		}

		metrics <- prometheus.MustNewConstMetric(c.successDesc, prometheus.GaugeValue, successValue, cisLabelValues...)

		if config.DefaultMutableFeatureGate.Enabled(feature.PolicyReport) {
			report := openreportsv1alpha1.Report{}
			if err := c.Get(ctx, client.ObjectKeyFromObject(&cis), &report); err != nil {
				c.Log.Error(err, "Failed to get Report", "namespace", cis.Namespace, "name", cis.Name)
				continue
			}

			severities := make(map[string]int32)

			patchStatuses := make(map[string]int32)

			for _, r := range report.Results {
				severities[reportSeverityCompat(r.Severity)]++
				if _, ok := r.Properties["fixedVersion"]; ok {
					patchStatuses["fixed"]++
				} else {
					patchStatuses["unfixed"]++
				}
			}

			for severity, count := range severities {
				issuesLabelValues[len(issuesLabelValues)-1] = severity
				metrics <- prometheus.MustNewConstMetric(c.issuesDesc, prometheus.GaugeValue, float64(count), issuesLabelValues...)
			}

			for patchStatus, count := range patchStatuses {
				patchStatusLabelValues[len(patchStatusLabelValues)-1] = patchStatus
				metrics <- prometheus.MustNewConstMetric(c.patchStatusDesc, prometheus.GaugeValue, float64(count), patchStatusLabelValues...)
			}

			continue
		}

		severities := cis.Status.VulnerabilitySummary.GetSeverityCount()
		for severity, count := range severities {
			issuesLabelValues[len(issuesLabelValues)-1] = severity
			metrics <- prometheus.MustNewConstMetric(c.issuesDesc, prometheus.GaugeValue, float64(count), issuesLabelValues...)
		}

		if cis.Status.VulnerabilitySummary != nil {
			patchStatusLabelValues[len(patchStatusLabelValues)-1] = "fixed"
			metrics <- prometheus.MustNewConstMetric(c.patchStatusDesc, prometheus.GaugeValue, float64(cis.Status.VulnerabilitySummary.FixedCount), patchStatusLabelValues...)

			patchStatusLabelValues[len(patchStatusLabelValues)-1] = "unfixed"
			metrics <- prometheus.MustNewConstMetric(c.patchStatusDesc, prometheus.GaugeValue, float64(cis.Status.VulnerabilitySummary.UnfixedCount), patchStatusLabelValues...)
		}
	}
}

// reportSeverityCompat is converting the severity label values to the currently used values
// We should probably consider a (metrics) breaking change in a follow-up PR to avoid this.
func reportSeverityCompat(reportSeverity openreportsv1alpha1.ResultSeverity) string {
	return strings.ToUpper(string(reportSeverity))
}

type cisMetricsLabelFunc func(cis stasv1alpha1.ContainerImageScan) string

type cisLabel struct {
	name  string
	value cisMetricsLabelFunc
}

type cisLabels []cisLabel

func (cl cisLabels) names(additionalNames ...string) []string {
	names := make([]string, 0, len(cl)+len(additionalNames))
	for _, l := range cl {
		names = append(names, l.name)
	}

	names = append(names, additionalNames...)

	return names
}

// Ensure ImageMetricsCollector is leader-election aware.
var _ manager.LeaderElectionRunnable = &ImageMetricsCollector{}
