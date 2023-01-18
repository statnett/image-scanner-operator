# Changelog

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
