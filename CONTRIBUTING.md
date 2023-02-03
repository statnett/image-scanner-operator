# Contributing

Following guidelines cover setting up dev environment, running, testing and deploying locally.

## Table of Contents

- [Set up your Development Environment](#set-up-your-development-environment)
- [Contribution Workflow](#contribution-workflow)
  - [Submitting an Issue](#submitting-an-issue)
  - [Submitting a Pull Request](#submitting-a-pull-request)
- [Custom Resource Definitions](#custom-resource-definitions)
  - [Generating code and manifests](#generating-code-and-manifests)
- [Deploy Image Scanner Operator](#deploy-image-scanner-operator)
  - [In Cluster](#in-cluster)

## Set up your Development Environment

1. Install Go

   The project requires [Go 1.19][go-download] or later. We also assume that you're familiar with
   Go's [GOPATH workspace][go-code] convention, and have the appropriate environment variables set.

1. Get the source code:

   ```shell
   git clone git@github.com:statnett/image-scanner-operator.git
   cd image-scanner-operator
   ```

## Contribution Workflow

## Submitting an Issue

Before you submit an issue, please search the issue tracker. An issue for your problem might already exist and the
discussion might inform you of workarounds readily available.

We want to fix all the issues as soon as possible, but before fixing a bug, we need to reproduce and confirm it.
To reproduce bugs, we require that you provide minimal reproduction.
Having a minimal reproducible scenario gives us a wealth of important information without going back and forth to
you with additional questions.

## Submitting a Pull Request

Before you submit your Pull Request (PR) consider the following guidelines:

1. Search [GitHub][github-pr] for an open or closed PR that relates
   to your submission. You don't want to duplicate existing efforts.

1. Be sure that an issue describes the problem you're fixing, or documents the design for the feature you'd like to add.
   Discussing the design upfront helps to ensure that we're ready to accept your work.

1. [Fork][image-scanner-repo] the repo.

1. In your forked repository, make your changes in a new git branch:

   ```shell
   git checkout -b my-fix-branch main
   ```

1. Create your patch, **including appropriate test cases**.

1. Run `make test` to run both unit and integrations (envtest) tests. [envtest][envtest] runs etcd and apiserver
   locally without the need for a real Kubernetes cluster. It helps to test the controller and the reconciliation logic.
   Note: Requires `ginkgo` binary to run the tests:
   ```shell
   go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.8.0
   ```

1. Optionally, run the e2e-tests. The e2e tests assumes that you have a working kubernetes cluster (e.g. kind or k3s cluster)
   and `KUBECONFIG` environment variable is pointing to that cluster configuration file. For example:

   ```shell
   export KUBECONFIG=~/.kube/config
   ```

   Note: The operator requires that some dependant software (at the moment [Prometheus Operator][prom-operator])
   is installed in the cluster. If not already present, it can be provisioned
   by running `make deploy-dependencies`.

   Run the e2e-tests by running `make e2e-test`.

1. Run `golangci-lint` to catch any linter errors.

   ```shell
   golangci-lint run
   ```

1. Commit your changes using a descriptive commit message.

   ```shell
   git commit
   ```

1. Push your branch to GitHub:

   ```shell
   git push origin my-fix-branch
   ```

1. In GitHub, send a pull request to `statnett:main`.

   PR title should be well written as that becomes the commit message when merging. We follow strict
   [Semantic release][semantic-release] to determine next semantic version number, generate a changelog and publish
   the release. Adherence to these conventions is a *MUST* because release notes are automatically generated from
   these messages.

   Single commit PRs are preferred. We follow [trunk based development][trunk-based-development] so larger changes
   must be broken down to multiple single commit PRs.

## Custom Resource Definitions

### Generating code and manifests

This project uses [`controller-gen`][controller-gen] to generate utility code and Kubernetes
manifests from source code and code markers. We currently generate:

- Custom Resource Definitions (CRD)
- RBACs
- Mandatory DeepCopy functions for a Go struct representing a CRD

This means that you should not try to modify any of these files directly, but instead change
the code and code markers. Our Makefile contains a target to ensure that all generated files
are up-to-date: So after doing modifications in code, affecting CRDs/RBAC, you should
run `make generate-all` to regenerate everything.

Our CI will verify that all generated is up-to-date.

Any change to the CRD structs, including nested structs, will probably modify the CRD.
This is also true for Go docs, as field/type doc becomes descriptions in CRDs.

When it comes to code markers added to the code, run `controller-gen -h` for detailed
reference (add more `h`'s to the command to get more details) or the
[markers documentation][markers-doc] for an overview.

We are trying to place the [RBAC markers][rbac-markers] close to the code that drives the
requirement for permissions. This could lead to the same, or similar, RBAC markers multiple
places in the code. This is how we want it to be, since it will allow us to track RBAC changes to
code changes. Any permission granted multiple times by markers will be deduplicated by controller-gen.

## Deploy Image Scanner Operator

By default, the operator is deployed to `image-scanner` namespace and monitors all the namespaces.

### In cluster

1. Install the image-scanner CRD via

   ```shell
   make install
   ```

1. Build and deploy the operator image into the kind/k3s cluster:

   ```shell
   make k3d-deploy
   ```

You can uninstall the operator with:

```shell
make undeploy
```

[go-download]: https://golang.org/dl/
[go-code]: https://golang.org/doc/code.html
[github-pr]: https://github.com/statnett/image-scanner-operator/pulls
[image-scanner-repo]: https://github.com/statnett/image-scanner-operator
[envtest]: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest
[prom-operator]: https://github.com/prometheus-operator/prometheus-operator
[semantic-release]: https://github.com/semantic-release/semantic-release
[trunk-based-development]: https://trunkbaseddevelopment.com/branch-by-abstraction/
[controller-gen]: https://book.kubebuilder.io/reference/controller-gen.html
[markers-doc]: https://book.kubebuilder.io/reference/markers.html
[rbac-markers]: https://book.kubebuilder.io/reference/markers/rbac.html
