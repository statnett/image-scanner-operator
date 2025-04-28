package config

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type Config struct {
	MetricsLabels                  []string       `mapstructure:"cis-metrics-labels"`
	ScanInterval                   time.Duration  `mapstructure:"scan-interval"`
	ScanJobNamespace               string         `mapstructure:"scan-job-namespace"`
	ScanJobServiceAccount          string         `mapstructure:"scan-job-service-account"`
	ScanJobTTLSecondsAfterFinished int32          `mapstructure:"scan-job-ttl-seconds-after-finished"`
	ScanNamespaces                 []string       `mapstructure:"namespaces"`
	ScanNamespaceExcludeRegexp     *regexp.Regexp `mapstructure:"scan-namespace-exclude-regexp"`
	ScanNamespaceIncludeRegexp     *regexp.Regexp `mapstructure:"scan-namespace-include-regexp"`
	ScanWorkloadResources          []string       `mapstructure:"scan-workload-resources"`
	SkipScanPodWaitingReasons      []string       `mapstructure:"skip-scan-pod-waiting-reasons"`
	TrivyImage                     string         `mapstructure:"trivy-image"`
	TrivyCommand                   TrivyCommand   `mapstructure:"trivy-command"`
	ActiveScanJobLimit             int            `mapstructure:"active-scan-job-limit"`
}

type TrivyCommand string

const (
	FilesystemTrivyCommand TrivyCommand = "filesystem"
	RootfsTrivyCommand     TrivyCommand = "rootfs"
)

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *TrivyCommand")

// UnmarshalText unmarshals text to a trivy command.
func (c *TrivyCommand) UnmarshalText(text []byte) error {
	if c == nil {
		return errUnmarshalNilLevel
	}

	if !c.unmarshalText(text) {
		return fmt.Errorf("unrecognized trivy command: %q", text)
	}

	return nil
}

func (c *TrivyCommand) unmarshalText(text []byte) bool {
	switch string(text) {
	case "filesystem":
		*c = FilesystemTrivyCommand
	case "rootfs":
		*c = RootfsTrivyCommand
	default:
		return false
	}

	return true
}
