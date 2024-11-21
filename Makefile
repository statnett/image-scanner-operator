
# Image URL to use all building/pushing image targets
IMG ?= "registry.dummy-domain.com/image-scanner/controller:dev"
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0
# Namespace to install operator into
K8S_NAMESPACE ?= image-scanner

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen k8s-client-gen ## Generate code required for K8s API and clients
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

GO_MODULE = $(shell go list -m)
API_DIRS = $(shell find api -mindepth 2 -type d | sed "s|^|$(shell go list -m)/|" | paste -sd " ")
.PHONY: k8s-client-gen
k8s-client-gen: applyconfiguration-gen
	@echo ">> generating internal/client/applyconfiguration..."
	$(APPLYCONFIGURATION_GEN) \
		--output-dir "internal/client/applyconfiguration" \
		--output-pkg "$(GO_MODULE)/internal/client/applyconfiguration" \
		$(API_DIRS)

.PHONY: wg-policy-client-gen
wg-policy-client-gen: applyconfiguration-gen
	@echo ">> generating internal/wg-policy/applyconfiguration..."
	$(APPLYCONFIGURATION_GEN) \
		--output-dir "internal/wg-policy/applyconfiguration" \
		--output-pkg "$(GO_MODULE)/internal/wg-policy/applyconfiguration" \
		"sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/reports.x-k8s.io/v1beta2" \
		"sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/wgpolicyk8s.io/v1alpha2"

.PHONY: wg-policy-crd-update
wg-policy-crd-update:
	curl -O --output-dir config/wg-policy/crd/ --remote-name-all \
		https://raw.githubusercontent.com/kubernetes-sigs/wg-policy-prototypes/master/policy-report/crd/v1beta2/reports.x-k8s.io_{clusterpolicyreports,policyreports}.yaml \
		https://raw.githubusercontent.com/kubernetes-sigs/wg-policy-prototypes/master/policy-report/crd/v1alpha2/wgpolicyk8s.io_{clusterpolicyreports,policyreports}.yaml

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

fmt-imports: gci ## Run gci against code.
	$(GCI) write --skip-generated -s standard -s default -s "prefix($(shell go list -m))" .

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -race ./... -coverprofile cover.out

test-reports: test ## Run tests and generate reports (no reports for now)

.PHONY: update-scan-log
update-scan-log: ## Update successful scan job pod log used in tests from template
	trivy image nginxinc/nginx-unprivileged@sha256:6da1811b094adbea1eb34c3e48fc2833b1a11a351ec7b36cc390e740a64fbae4 \
		--offline-scan --severity CRITICAL,HIGH --quiet --format template \
		--template @$(shell pwd)/internal/trivy/templates/scan-report.json.tmpl \
		> internal/controller/stas/testdata/scan-job-successful/successful-scan-job-pod.log.json

.PHONY: go-mod-tidy
go-mod-tidy: ## Run go mod tidy against code.
	go mod tidy

.PHONY: generate-all
generate-all: manifests generate fmt fmt-imports go-mod-tidy ## Ensure all generated files are up-to-date

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} --build-arg GOPROXY=${GOPROXY} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply --server-side -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/e2e-test | kubectl apply --server-side -f -
	for w in statefulset/trivy deployment/image-scanner-controller-manager; do \
		if ! kubectl rollout status $$w -n $(K8S_NAMESPACE) --timeout=2m; then \
			kubectl get events -n $(K8S_NAMESPACE); \
			kubectl logs -n $(K8S_NAMESPACE) $$w --tail -1; \
			exit 1; \
		fi \
  	done

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/e2e-test | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-dependencies
deploy-dependencies: kustomize ## Install operator dependent software not part of standard K8s
	$(KUSTOMIZE) build config/deploy-dependencies | kubectl apply --server-side -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
APPLYCONFIGURATION_GEN ?= $(LOCALBIN)/applyconfiguration-gen
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
# renovate: datasource=go depName=github.com/daixiang0/gci
GCI_VERSION ?= v0.13.5

## Tool Versions
# renovate: datasource=go depName=sigs.k8s.io/kustomize/kustomize/v5
KUSTOMIZE_VERSION ?= v5.5.0
# renovate: datasource=go depName=github.com/kubernetes/code-generator
CODE_GENERATOR_VERSION ?= v0.31.3
# renovate: datasource=go depName=sigs.k8s.io/controller-tools
CONTROLLER_TOOLS_VERSION ?= v0.16.5

.PHONY: applyconfiguration-gen
applyconfiguration-gen: $(APPLYCONFIGURATION_GEN) ## Download applyconfiguration-gen locally if necessary.
$(APPLYCONFIGURATION_GEN): $(LOCALBIN)
	# FIXME: applyconfiguration-gen does not currently support any flag for obtaining version
	test -s $(LOCALBIN)/applyconfiguration-gen || \
	GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/applyconfiguration-gen@$(CODE_GENERATOR_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

GCI = $(LOCALBIN)/gci
gci: $(GCI) ## Download gci locally if necessary.
$(GCI):
	GOBIN=$(LOCALBIN) go install github.com/daixiang0/gci@${GCI_VERSION}

##@ End-to-end (e2e) Testing

e2e-test: ## Run e2e tests
	chainsaw test --config test/e2e-config/.chainsaw.yaml \
	--test-dir test/e2e/scenario \
	--test-dir test/e2e/workload-scan

##@ K3d

K3D_CLUSTER ?= image-scanner

k3d-deploy: deploy-dependencies k3d-image-import deploy ## Build and deploy operator to K3d cluster

k3d-image-import: docker-build ## Imports controller image into k3d cluster
	k3d image import --cluster "${K3D_CLUSTER}" "${IMG}"

k3d-cluster-create: ## Create a default K3d cluster
	k3d cluster create "${K3D_CLUSTER}" --config test/e2e-config/k3d-config.yml

k3d-cluster-delete: ## Delete the default K3d cluster
	k3d cluster delete "${K3D_CLUSTER}"
