# Changelog

## [0.2.2](https://github.com/statnett/image-scanner-operator/compare/v0.2.1...v0.2.2) (2023-02-03)


### Bug Fixes

* don't override user logging configuration ([#178](https://github.com/statnett/image-scanner-operator/issues/178)) ([999ced3](https://github.com/statnett/image-scanner-operator/commit/999ced36d76bd3be2a2152f201cfc4b6690f04dc))


### Dependency Updates

* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.2 to 0.1.3 ([#179](https://github.com/statnett/image-scanner-operator/issues/179)) ([e9ff6bf](https://github.com/statnett/image-scanner-operator/commit/e9ff6bf68dad9455fe3488283e44c865352e5497))

## [0.2.1](https://github.com/statnett/image-scanner-operator/compare/v0.2.0...v0.2.1) (2023-02-02)


### Bug Fixes

* do not log DEBUG lines by default ([#165](https://github.com/statnett/image-scanner-operator/issues/165)) ([5ab234a](https://github.com/statnett/image-scanner-operator/commit/5ab234a57fa0e3f1ad6948133b6c8f8aea020ea4))
* make logger configurable again ([#166](https://github.com/statnett/image-scanner-operator/issues/166)) ([45712ce](https://github.com/statnett/image-scanner-operator/commit/45712cee7803039e2095e6f15e4934c763d5dca8))

## [0.2.0](https://github.com/statnett/image-scanner-operator/compare/v0.1.1...v0.2.0) (2023-02-01)


### ⚠ BREAKING CHANGES

* move code out of main package ([#141](https://github.com/statnett/image-scanner-operator/issues/141))
* **admin:** kustomizeable configmap for scan job default configuration ([#110](https://github.com/statnett/image-scanner-operator/issues/110))

### Features

* add Openshift anyuid SCC role binding ([#160](https://github.com/statnett/image-scanner-operator/issues/160)) ([03baa7d](https://github.com/statnett/image-scanner-operator/commit/03baa7dc05080e2ea1cdf0467d3b047d40986328))
* **admin:** kustomizeable configmap for scan job default configuration ([#110](https://github.com/statnett/image-scanner-operator/issues/110)) ([d51e2a9](https://github.com/statnett/image-scanner-operator/commit/d51e2a977496dbbfd7ae22c3283ac3957bbe94d4))
* **cli:** make regexp for excluded namespace configurable ([#79](https://github.com/statnett/image-scanner-operator/issues/79)) ([9a50838](https://github.com/statnett/image-scanner-operator/commit/9a50838dca99c5dcc59b36f1e4da82306b0cf13b))
* **cli:** make regexp for included namespace configurable ([#157](https://github.com/statnett/image-scanner-operator/issues/157)) ([4b145ca](https://github.com/statnett/image-scanner-operator/commit/4b145caee2d1efa7c456d2216c440bc041e69c09))


### Bug Fixes

* init temporary logger to print errors during startup ([#143](https://github.com/statnett/image-scanner-operator/issues/143)) ([d9a6134](https://github.com/statnett/image-scanner-operator/commit/d9a613438f17936366a8b21b24d26c70cb6dfc6c))


### refactor

* move code out of main package ([#141](https://github.com/statnett/image-scanner-operator/issues/141)) ([b495881](https://github.com/statnett/image-scanner-operator/commit/b4958815c5260e1971147d319a0dc680626ba13b))


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.7.0 to 2.7.1 ([#153](https://github.com/statnett/image-scanner-operator/issues/153)) ([73f3e81](https://github.com/statnett/image-scanner-operator/commit/73f3e8111e5d8af71d6091f8610ff36e6c675429))
* **deps:** bump github.com/onsi/ginkgo/v2 from 2.7.1 to 2.8.0 ([#159](https://github.com/statnett/image-scanner-operator/issues/159)) ([9159cfc](https://github.com/statnett/image-scanner-operator/commit/9159cfc11170615466820fbafce304553f733172))
* **deps:** bump github.com/onsi/gomega from 1.25.0 to 1.26.0 ([#122](https://github.com/statnett/image-scanner-operator/issues/122)) ([bcfbd53](https://github.com/statnett/image-scanner-operator/commit/bcfbd53171599a0f60d697c26bcbc26cd60d5d90))
* **deps:** bump github.com/spf13/viper from 1.14.0 to 1.15.0 ([#107](https://github.com/statnett/image-scanner-operator/issues/107)) ([7d541cb](https://github.com/statnett/image-scanner-operator/commit/7d541cb2d63b3a1c233ad9a767b8cf3e5d625753))
* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.1 to 0.1.2 ([#161](https://github.com/statnett/image-scanner-operator/issues/161)) ([2c7ce33](https://github.com/statnett/image-scanner-operator/commit/2c7ce33ceb06807b54c7a9d4afbef2c5eb73d781))
* **deps:** bump k8s.io/api from 0.26.0 to 0.26.1 ([#104](https://github.com/statnett/image-scanner-operator/issues/104)) ([1d17f1e](https://github.com/statnett/image-scanner-operator/commit/1d17f1efe5bb49d15bb33eb9bb94f31bd3accb4d))
* **deps:** bump k8s.io/apimachinery from 0.26.0 to 0.26.1 ([#105](https://github.com/statnett/image-scanner-operator/issues/105)) ([74ac40e](https://github.com/statnett/image-scanner-operator/commit/74ac40e3b992d7d0b2c8999eb3a00d72bba520c0))
* **deps:** bump k8s.io/client-go from 0.26.0 to 0.26.1 ([#106](https://github.com/statnett/image-scanner-operator/issues/106)) ([aca2d41](https://github.com/statnett/image-scanner-operator/commit/aca2d411809bc068ed4a7abaf635cf153d07beca))
* **deps:** bump k8s.io/klog/v2 from 2.80.1 to 2.90.0 ([#123](https://github.com/statnett/image-scanner-operator/issues/123)) ([42c7eca](https://github.com/statnett/image-scanner-operator/commit/42c7eca1de1b35b0d54fd78b3898f24bd1d03320))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.14.1 to 0.14.2 ([#152](https://github.com/statnett/image-scanner-operator/issues/152)) ([e3e0a51](https://github.com/statnett/image-scanner-operator/commit/e3e0a51573e70752e8fbd0953fe8b6e5898b4bab))

## [0.1.1](https://github.com/statnett/image-scanner-operator/compare/v0.1.0...v0.1.1) (2023-01-18)


### Bug Fixes

* add workload labels to the podTemplate in the scan Job ([#83](https://github.com/statnett/image-scanner-operator/issues/83)) ([b4327fd](https://github.com/statnett/image-scanner-operator/commit/b4327fdbef45ae88c114ad7c42c569e0640ce182))
* obtain image tags from pod container spec ([#86](https://github.com/statnett/image-scanner-operator/issues/86)) ([93c183f](https://github.com/statnett/image-scanner-operator/commit/93c183f73dff4d507cbe28a71109ac55fb161559))
* update CIS status with errors decoding scan report ([#90](https://github.com/statnett/image-scanner-operator/issues/90)) ([27e5c51](https://github.com/statnett/image-scanner-operator/commit/27e5c51b45673e1fff0bd597a60473472f21680f))


### Dependency Updates

* **deps:** bump github.com/onsi/gomega from 1.24.2 to 1.25.0 ([#88](https://github.com/statnett/image-scanner-operator/issues/88)) ([74860c2](https://github.com/statnett/image-scanner-operator/commit/74860c2530407241e089b7f44f5565a2759bacbb))

## [0.1.0](https://github.com/statnett/image-scanner-operator/compare/v0.0.1...v0.1.0) (2023-01-17)


### Features

* add workload labels on the scan job ([#80](https://github.com/statnett/image-scanner-operator/issues/80)) ([0faade8](https://github.com/statnett/image-scanner-operator/commit/0faade88ca218ba699452a448080af0af8bd456d))


### Bug Fixes

* add GenerationChangedPredicate to CIS controller ([#66](https://github.com/statnett/image-scanner-operator/issues/66)) ([9d6372b](https://github.com/statnett/image-scanner-operator/commit/9d6372b5336d907ac4875e8f1caaa984d9942f83))
* **cli:** should print help when requested ([#62](https://github.com/statnett/image-scanner-operator/issues/62)) ([f7dd4d3](https://github.com/statnett/image-scanner-operator/commit/f7dd4d39fd19e173dd8b54bce0f0ad1b983d6f63))
* don't use generateName when creating jobs ([#48](https://github.com/statnett/image-scanner-operator/issues/48)) ([d7c8185](https://github.com/statnett/image-scanner-operator/commit/d7c818552fdbb6442caebf8b06bf1865a9df4c66))

## [0.0.1](https://github.com/statnett/image-scanner-operator/compare/v0.0.0...v0.0.1) (2023-01-12)


### Bug Fixes

* **ci:** use if condition to check rollout status ([#24](https://github.com/statnett/image-scanner-operator/issues/24)) ([8f8f173](https://github.com/statnett/image-scanner-operator/commit/8f8f173fb86857eb7d5906943d627c7cba69f3ea))
* **cli:** print help text without the err message ([#29](https://github.com/statnett/image-scanner-operator/issues/29)) ([7bcbc27](https://github.com/statnett/image-scanner-operator/commit/7bcbc270a7655b3a52a1d0075b975f89e0b254f7))
* **log:** initialize klog to get consistent log format ([#42](https://github.com/statnett/image-scanner-operator/issues/42)) ([3d9cbdf](https://github.com/statnett/image-scanner-operator/commit/3d9cbdf9c32ea9cb10ef9c24f78c5fad0f22f599))


### Dependency Updates

* **deps:** bump aquasecurity/trivy from 0.35.0 to 0.36.1 ([#40](https://github.com/statnett/image-scanner-operator/issues/40)) ([aedfb3a](https://github.com/statnett/image-scanner-operator/commit/aedfb3a786842f7573cea73295e044ff570d6b7f))
* **deps:** bump github.com/onsi/ginkgo/v2 from 2.6.1 to 2.7.0 ([#23](https://github.com/statnett/image-scanner-operator/issues/23)) ([a7edc92](https://github.com/statnett/image-scanner-operator/commit/a7edc9250d2b2ef82c7bbc53853a41f1fdfbce7c))
* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.0 to 0.1.1 ([#3](https://github.com/statnett/image-scanner-operator/issues/3)) ([92ef9c0](https://github.com/statnett/image-scanner-operator/commit/92ef9c03ea11a399ffcc12eb79f0e7917ab60b41))
* **deps:** bump github.com/vektra/mockery/v2 from 2.15.0 to 2.16.0 ([#2](https://github.com/statnett/image-scanner-operator/issues/2)) ([7011dc9](https://github.com/statnett/image-scanner-operator/commit/7011dc9b14019faeffbaba145279a7626bd65c57))
