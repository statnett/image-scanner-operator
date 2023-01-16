package trivy

import (
	"embed"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/pointer"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/hash"
	"github.com/statnett/image-scanner-operator/pkg/operator"
)

const (
	FsScanSharedVolumeMountPath   = "/var/run/image-scanner"
	FsScanSharedVolumeName        = "image-scanner"
	FsScanTrivyBinaryPath         = FsScanSharedVolumeMountPath + "/trivy"
	JobNameSpecHashPartLength     = 5
	KubernetesJobNameMaxLength    = validation.DNS1123LabelMaxLength
	KubernetesLabelValueMaxLength = validation.DNS1123LabelMaxLength
	ScanJobContainerName          = "scan-image"
	ScanJobTimeout                = 1 * time.Hour
	TempVolumeName                = "tmp"
	TempVolumeMountPath           = "/tmp"
)

var (
	//go:embed templates/*
	templatesFS    embed.FS
	reportTemplate string
)

func init() {
	bytes, err := templatesFS.ReadFile("templates/scan-report.json.tmpl")
	if err != nil {
		panic(err)
	}

	reportTemplate = string(bytes)
}

type ImageScanJobBuilder interface {
	OnPreferredNodes(nodeNames ...string) ImageScanJobBuilder
	ForCIS(cis *stasv1alpha1.ContainerImageScan) (*batchv1.Job, error)
}

func NewImageScanJob(config operator.Config) ImageScanJobBuilder {
	return &filesystemScanJobBuilder{
		Config: config,
	}
}

type filesystemScanJobBuilder struct {
	operator.Config
	preferredNodeNames []string
}

func (f *filesystemScanJobBuilder) ForCIS(cis *stasv1alpha1.ContainerImageScan) (*batchv1.Job, error) {
	job, err := f.newImageScanJob(cis.Spec)
	if err != nil {
		return job, err
	}

	job.Namespace = f.ScanJobNamespace
	job.Name = scanJobName(cis)
	job.Labels = map[string]string{
		stasv1alpha1.LabelK8sAppName:                  stasv1alpha1.AppNameTrivy,
		stasv1alpha1.LabelK8SAppManagedBy:             stasv1alpha1.AppNameImageScanner,
		stasv1alpha1.LabelStatnettControllerNamespace: cis.Namespace,
		stasv1alpha1.LabelStatnettControllerUID:       string(cis.UID),
		stasv1alpha1.LabelStatnettWorkloadKind:        cis.Spec.Workload.Kind,
		stasv1alpha1.LabelStatnettWorkloadName:        workloadLabelName(cis),
		stasv1alpha1.LabelStatnettWorkloadNamespace:   cis.Namespace,
	}

	return job, nil
}

func workloadLabelName(cis *stasv1alpha1.ContainerImageScan) string {
	if len(cis.Spec.Workload.Name) > KubernetesLabelValueMaxLength {
		return cis.Spec.Workload.Name[0 : KubernetesLabelValueMaxLength-1]
	} else {
		return cis.Spec.Workload.Name
	}
}

func scanJobName(cis *stasv1alpha1.ContainerImageScan) string {
	hashPart := hash.NewString(cis.Spec, cis.Namespace)[0:JobNameSpecHashPartLength]
	nameFn := func(cisName string) string {
		return fmt.Sprintf("%s-%s", cisName, hashPart)
	}

	name := nameFn(cis.Name)
	if len(name) > KubernetesJobNameMaxLength {
		shortenCISName := cis.Name[0 : len(cis.Name)-(len(name)-KubernetesJobNameMaxLength)]
		name = nameFn(shortenCISName)
	}

	return name
}

func (f *filesystemScanJobBuilder) newImageScanJob(spec stasv1alpha1.ContainerImageScanSpec) (*batchv1.Job, error) {
	job := &batchv1.Job{}

	container, err := f.container(spec)
	if err != nil {
		return nil, err
	}

	job.Spec.Template.Labels = map[string]string{
		stasv1alpha1.LabelK8sAppName:      stasv1alpha1.AppNameTrivy,
		stasv1alpha1.LabelK8SAppManagedBy: stasv1alpha1.AppNameImageScanner,
	}
	job.Spec.Template.Spec.InitContainers = []corev1.Container{f.initContainer()}
	job.Spec.Template.Spec.Containers = []corev1.Container{container}
	job.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: FsScanSharedVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: TempVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	job.Spec.Parallelism = pointer.Int32(1)
	job.Spec.Completions = pointer.Int32(1)
	job.Spec.ActiveDeadlineSeconds = pointer.Int64(int64(ScanJobTimeout.Seconds()))
	job.Spec.BackoffLimit = pointer.Int32(3)
	job.Spec.TTLSecondsAfterFinished = pointer.Int32(7200)
	job.Spec.Template.Spec.ServiceAccountName = f.ScanJobServiceAccount

	if len(f.preferredNodeNames) > 0 {
		terms := make([]corev1.PreferredSchedulingTerm, len(f.preferredNodeNames))
		for i, nodeName := range f.preferredNodeNames {
			terms[i] = corev1.PreferredSchedulingTerm{
				Weight: 100,
				Preference: corev1.NodeSelectorTerm{
					MatchFields: []corev1.NodeSelectorRequirement{{
						Key:      "metadata.name",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{nodeName},
					}},
				},
			}
		}

		job.Spec.Template.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: terms,
			},
		}
	}

	job.Spec.Template.Spec.AutomountServiceAccountToken = pointer.Bool(false)
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure

	return job, nil
}

func (f *filesystemScanJobBuilder) OnPreferredNodes(nodeNames ...string) ImageScanJobBuilder {
	f.preferredNodeNames = nodeNames
	return f
}

func (f *filesystemScanJobBuilder) container(spec stasv1alpha1.ContainerImageScanSpec) (corev1.Container, error) {
	container := corev1.Container{}

	canonical, err := spec.Image.Canonical()
	if err != nil {
		return container, err
	}

	container.Name = ScanJobContainerName
	container.Image = canonical.String()
	container.Command = []string{FsScanTrivyBinaryPath}
	container.Args = []string{
		"filesystem",
		"/",
	}
	container.Env = []corev1.EnvVar{
		{Name: "HOME", Value: TempVolumeMountPath},
		{Name: "TRIVY_OFFLINE_SCAN", Value: "true"},
		{Name: "TRIVY_SECURITY_CHECKS", Value: "vuln"},
		{Name: "TRIVY_CACHE_DIR", Value: TempVolumeMountPath},
		{Name: "TRIVY_SERVER", Value: f.TrivyServer},
		{Name: "TRIVY_QUIET", Value: "true"},
		{Name: "TRIVY_FORMAT", Value: "template"},
		{Name: "TRIVY_TEMPLATE", Value: reportTemplate},
		{Name: "TRIVY_TIMEOUT", Value: ScanJobTimeout.String()},
	}

	if spec.MinSeverity != nil {
		minSeverity, err := stasv1alpha1.NewSeverity(*spec.MinSeverity)
		if err != nil {
			return container, err
		}

		var severityNames []string

		for severity := minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
			severityNames = append(severityNames, severity.String())
		}

		envVar := corev1.EnvVar{
			Name:  "TRIVY_SEVERITY",
			Value: strings.Join(severityNames, ","),
		}
		container.Env = append(container.Env, envVar)
	}

	if pointer.BoolDeref(spec.IgnoreUnfixed, false) {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "TRIVY_IGNORE_UNFIXED",
			Value: "true",
		})
	}

	container.Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("500M"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100M"),
		},
	}
	container.SecurityContext = &corev1.SecurityContext{
		Privileged:               pointer.Bool(false),
		AllowPrivilegeEscalation: pointer.Bool(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"all"},
		},
		ReadOnlyRootFilesystem: pointer.Bool(true),
		RunAsUser:              pointer.Int64(0),
	}
	container.TerminationMessagePolicy = corev1.TerminationMessageFallbackToLogsOnError
	container.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      FsScanSharedVolumeName,
			MountPath: FsScanSharedVolumeMountPath,
		},
		{
			Name:      TempVolumeName,
			MountPath: TempVolumeMountPath,
		},
	}
	container.WorkingDir = TempVolumeMountPath

	return container, nil
}

func (f *filesystemScanJobBuilder) initContainer() corev1.Container {
	return corev1.Container{
		Name:                     "trivy",
		Image:                    f.TrivyImage,
		ImagePullPolicy:          corev1.PullIfNotPresent,
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
		Command: []string{
			"cp",
			"-v",
			"/usr/local/bin/trivy",
			FsScanTrivyBinaryPath,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("10m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      FsScanSharedVolumeName,
				ReadOnly:  false,
				MountPath: FsScanSharedVolumeMountPath,
			},
		},
	}
}
