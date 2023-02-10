package config

import (
	"regexp"
	"time"
)

type Config struct {
	MetricsLabels              []string       `mapstructure:"cis-metrics-labels"`
	ScanInterval               time.Duration  `mapstructure:"scan-interval"`
	ScanJobNamespace           string         `mapstructure:"scan-job-namespace"`
	ScanJobServiceAccount      string         `mapstructure:"scan-job-service-account"`
	ScanNamespaces             []string       `mapstructure:"namespaces"`
	ScanNamespaceExcludeRegexp *regexp.Regexp `mapstructure:"scan-namespace-exclude-regexp"`
	ScanNamespaceIncludeRegexp *regexp.Regexp `mapstructure:"scan-namespace-include-regexp"`
	ScanWorkloadResources      []string       `mapstructure:"scan-workload-resources"`
	TrivyImage                 string         `mapstructure:"trivy-image"`
}
