# Image Scanner Operator

[![Conventional Commits][conventional-commits-img]][conventional-commits]
[![Go Report Card][report-card-img]][report-card]
[![CodeQL][CodeQL-img]][CodeQL]

Image Scanner is a Kubernetes Operator supporting detection of vulnerabilities
in running container images.

Kubernetes clusters run containers created from a great diversity of container
images with several origins.
Some images are custom for an in-house application, while others are pulled
from internal/external registries.
It is considered best practice to scan images in container image build
pipelines, to provide developers with early feedback on potential
vulnerabilities.

While this is good, it is not enough:

Software vulnerabilities are typically detected after the software is available
for use, and some applications will typically "just run" in their runtime
environment - with little/no maintenance.
And some users might not use pipelines at all or use pipelines without image
scanning enabled.

Running applications on container images with vulnerabilities _might_ represent
an unacceptable threat.
To mend this, we want **a mechanism to identify vulnerabilities
in running container images**.
This could then be used by developers, system administrators, platform
administrators and security officers
to take action when vulnerabilities are detected.

Actions could for instance be:

- upgrade to a patched version of the vulnerable software component
- conclude that the vulnerability is not relevant
- conclude that the vulnerability represents an acceptable risk
- shut down the application

There seem to be quite a lot of companies/communities providing similar
software. Some key features of this operator are:

- **Can be configured to scan container images for any
  [Kubernetes Workload](https://kubernetes.io/docs/concepts/workloads/)**,
  including custom resources. The only requirement is that the resource
  MUST
  [own](https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/)
  Pods - either directly or indirectly.
- **Container images are scanned based on immutable image sha256 digests**
  obtained from
  [Container Runtime Interface (CRI)](https://kubernetes.io/docs/concepts/architecture/cri/),
  not from potentially mutable image tags.
- To avoid unnecessary image pulls, the image scan workload is preferred
  scheduled on the same node(s) as the workload to be scanned.
- Since the image to scan should already be present in the node container
  registry, **the image scan workload does not have to bother with image pull secrets for
  private images**, exploiting a
  [Kubernetes "bug"](https://github.com/kubernetes/kubernetes/issues/18787).

## Description

// TODO(user): An in-depth paragraph about your project and overview of use

### Custom resources

The Image Scanner operator currently defines a single user-facing Custom
Resource Definition (CRD), [ContainerImageScan][CIS-CRD] (CIS), that represents the
Kubernetes API for runtime image scanning of workload container images.
See [stas_v1alpha1_containerimagescan.yaml][CIS-example] for a (simplified)
example of a CIS resource.

The CIS resource `.spec` specifies the container image to scan and some
additional workload metadata, and the image scan result is added/updated
in `.status` by the `ContainerImageScan` controller.

CIS resources should not be edited by standard users, as the `Workload`
controller will create CIS resources from running pods. And the standard
Kubernetes garbage collector deletes the obsolete CIS resources when the
owning pods are gone.

A user can influence the image scanning process by adding annotations to pods.
The set of annotations is currently limited, but more might be added in the
future:

| Pod annotation key                         | Default value | Description                                                                                                           |
|--------------------------------------------|:--------------|:----------------------------------------------------------------------------------------------------------------------|
| `image-scanner.statnett.no/ignore-unfixed` | `"false"`     | If set to `"true"`, the Image Scanner will ignore any detected vulnerability that can't be fix by updating package(s) |

### Supported features

- Namespaced container image scan API (custom resource): `ContainerImageScan` (CIS)
- A CIS resource contains details and a summary of detected vulnerabilities
- Users can identify the owning/controlling resource (workload) of the scanned container image
- All container images are rescanned regularly with configurable interval
- Provides vulnerability summary metrics from CIS to enable dashboards/alerts
- Any user with access to a namespace is allowed to _view_ CIS
- Cluster-scoped operator that operates in configured namespaces with an
  include/block list of namespaces that should be scanned
- Supports any type of workload (also CRDs) by configuration
- Scanning workload images from private/authenticated image registries

### Future improvements

- Push-based feedback to stakeholders
- Enable users to ignore/suppress vulnerabilities
- Perform actions from detected vulnerabilities

### Eschewed Features

- Produce metrics for vulnerability details
- Historical container image scans
- Helm chart installation

## Getting Started

You’ll need a Kubernetes cluster to run the Image Scanner.
You can use [KIND](https://sigs.k8s.io/kind) or [k3s](https://k3s.io/)
to get a local cluster for testing.

We currently only support installation using
[Kustomize](https://kustomize.io/), either as a standalone tool,
using the kustomize (`-k`) feature in recent versions of `kubectl`
or any GitOps tool with support for Kustomize.

### Install

Since you probably want to adjust the default configuration (and/or have
multiple clusters), we suggest you start by creating a kustomize overlay using
the Image Scanner default kustomization as a
[remote directory](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/remoteBuild.md#remote-directories)
base. Your initial kustomization.yaml could be as simple as:

<!-- x-release-please-start-version -->
```yaml
resources:
  - https://github.com/statnett/image-scanner-operator?ref=v0.8.45
```
<!-- x-release-please-end -->

If you have multiple clusters, you should create one
[variant](https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#variant)
overlay per cluster.

To install (or update) the operator into your cluster run:

```sh
kubectl apply --server-side -k <overlay-directory>
```

### Configure

The Image Scanner operator is highly configurable and supports numerous
flags with corresponding environment variables. To get an overview over
all that can be configured, the easiest is to use the CLI:

```sh
docker run ghcr.io/statnett/image-scanner-operator --help
```

You can also use environment variables for configuration, but a flag takes
precedence over the corresponding environment variable.
Environment variable names can be deduced from flags by upper-casing and
replacing the `-` delimiter with `_`.

Since we use kustomize to install the operator, the easiest is
to customize the environment variables provided
from the ConfigMap in the default Image Scanner configuration using a
[configMapGenerator](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/configmapgenerator/).

```yaml
configMapGenerator:
  - name: image-scanner-config
    behavior: merge
    literals:
      - CIS_METRICS_LABELS=app.kubernetes.io/name
      - SCAN_INTERVAL=24h
```

This example will override the default configuration to:

- add metric labels with values obtained from `app.kubernetes.io/name` Pod labels
- rescan workload images with an interval of 24 hours

#### Configure Trivy scan jobs

A workload container image is scanned by scheduling a Kubernetes _Job_ running
on the scan target container image. The image is scanned using the
[`trivy filesystem`](https://aquasecurity.github.io/trivy/latest/docs/references/configuration/cli/trivy_filesystem/)
command inside the job's container.

The `trivy filesystem` scan command can be customized by modifying the
`trivy-job-config` _ConfigMap_. All entries in the _ConfigMap_ are mounted
as environment variables with the `TRIVY_` prefix - which will allow them
to be picked up by Trivy. Example:

```yaml
  - name: trivy-job-config
    namespace: image-scanner
    behavior: merge
    literals:
      - DB_REPOSITORY=<company-ghcr-registry-proxy>/aquasecurity/trivy-db
      - JAVA_DB_REPOSITORY=<company-ghcr-registry-proxy>/aquasecurity/trivy-java-db
      - OFFLINE_SCAN=true # enabling offline mode for air-gapped environments
```

### Upgrade

At this early stage, we might introduce breaking changes. But when we do,
the breaking changes should be highlighted in the changelog and release notes.

We might also do breaking changes in the CRD(s), without adding the complexity
of conversion webhooks. So if you experience any issue when upgrading to a
newer version, please try to reinstall the operator as a first step.

### Uninstall

To uninstall the operator, just use `kubectl` to delete all resources produced
by the overlay used when [installing](#install) the operator:

```sh
kubectl delete -k --ignore-not-found=true <overlay-directory>
```

### How it works

This project aims to follow the Kubernetes
[Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses
[Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources
until the desired state is reached.

Image Scanner consists of three controllers that coordinates
scanning of running container images as illustrated in the diagrams below.

The container image scan Kubernetes API is materialized
by the `ContainerImageScan` custom resources providing
an eventually consistent image scanning result in its
status.

The actual vulnerability scan of a container image and the vulnerability
database is provided by an external service to the operator.

Using a simple `Pod`, with a single container, as an example:

1. Create a `ContainerImageScan` when the immutable image reference is available in the `Pod` status.
2. Create a scan `Job` from immutable image reference in the `ContainerImageScan` spec.
3. When a scan `Job` is completed, read the scan result from pod log of the scan `Job`,
   and update the `ContainerImageScan` status.
4. When the `Pod` is deleted, the `ContainerImageScan` is garbage collected.

![Image scanner component diagram](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/statnett/image-scanner-operator/main/docs/operator-component.puml))
![Scan image sequence diagram](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/statnett/image-scanner-operator/main/docs/scan-sequence.puml)

## Contributing

We would love your feedback on any aspect of the Image Scanner!
Feel free to open issues for things you think can be improved.
Or/and open a PR (better) to show how we can improve.

See [Contributing](CONTRIBUTING.md) for information about setting up
your local development environment, and the contribution workflow expected.

Please ensure that you are following our [Code Of Conduct](CODE_OF_CONDUCT.md)
during any interaction with the community.

## License

Licensed under the [MIT License](LICENSE).

[CodeQL-img]: https://github.com/statnett/image-scanner-operator/actions/workflows/codeql.yml/badge.svg?branch=main
[CodeQL]: https://github.com/statnett/image-scanner-operator/actions/workflows/codeql.yml?query=branch%3Amain
[conventional-commits-img]: https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white
[conventional-commits]: https://conventionalcommits.org
[report-card-img]: https://goreportcard.com/badge/github.com/statnett/image-scanner-operator
[report-card]: https://goreportcard.com/report/github.com/statnett/image-scanner-operator
[CIS-CRD]: https://doc.crds.dev/github.com/statnett/image-scanner-operator/stas.statnett.no/ContainerImageScan/v1alpha1
[CIS-example]: config/samples/stas_v1alpha1_containerimagescan.yaml
