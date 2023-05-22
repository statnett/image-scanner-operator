# Changelog

## [0.5.7](https://github.com/statnett/image-scanner-operator/compare/v0.5.6...v0.5.7) (2023-05-22)


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.9.4 to 2.9.5 ([#378](https://github.com/statnett/image-scanner-operator/issues/378)) ([43ce334](https://github.com/statnett/image-scanner-operator/commit/43ce3345310005c7ae170fefc770a6dfc626a2c0))
* **deps:** bump github.com/onsi/gomega from 1.27.6 to 1.27.7 ([#382](https://github.com/statnett/image-scanner-operator/issues/382)) ([328ae39](https://github.com/statnett/image-scanner-operator/commit/328ae39517a87149a1e9f6b7a59fb9ab3b8779ff))
* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.8 to 0.1.9 ([#376](https://github.com/statnett/image-scanner-operator/issues/376)) ([fc97b58](https://github.com/statnett/image-scanner-operator/commit/fc97b5893984da8c421588224e3b20c4e061c859))
* **deps:** bump github.com/stretchr/testify from 1.8.2 to 1.8.3 ([#383](https://github.com/statnett/image-scanner-operator/issues/383)) ([5486fee](https://github.com/statnett/image-scanner-operator/commit/5486fee6a7f719c58836ec3b15bb82c2f3add731))
* **deps:** bump github.com/vektra/mockery/v2 from 2.26.1 to 2.27.1 ([#375](https://github.com/statnett/image-scanner-operator/issues/375)) ([e3b1551](https://github.com/statnett/image-scanner-operator/commit/e3b1551e61f2014813c5ec69d84bca77db331e30))

## [0.5.6](https://github.com/statnett/image-scanner-operator/compare/v0.5.5...v0.5.6) (2023-05-12)


### Bug Fixes

* revert change in scan job pod error handling ([#371](https://github.com/statnett/image-scanner-operator/issues/371)) ([6bc7359](https://github.com/statnett/image-scanner-operator/commit/6bc7359dc2051bc6fe1a26179ce3633e31707d2d))


### Dependency Updates

* **deps:** bump github.com/distribution/distribution from 2.8.1+incompatible to 2.8.2+incompatible ([#374](https://github.com/statnett/image-scanner-operator/issues/374)) ([9834d9f](https://github.com/statnett/image-scanner-operator/commit/9834d9fd44b5c36d7bb15ee06de64fdea36a2268))
* **deps:** bump github.com/docker/distribution from 2.8.1+incompatible to 2.8.2+incompatible ([#373](https://github.com/statnett/image-scanner-operator/issues/373)) ([d869eaf](https://github.com/statnett/image-scanner-operator/commit/d869eaf8c5381f3c132c08a9c54978105e58b4ba))

## [0.5.5](https://github.com/statnett/image-scanner-operator/compare/v0.5.4...v0.5.5) (2023-05-10)


### Bug Fixes

* don't consider evicted pods when looking for scan job pods ([#369](https://github.com/statnett/image-scanner-operator/issues/369)) ([3b0578b](https://github.com/statnett/image-scanner-operator/commit/3b0578bccaabeb2f3eb3d84b419d7eb9958cd719))

## [0.5.4](https://github.com/statnett/image-scanner-operator/compare/v0.5.3...v0.5.4) (2023-05-04)


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.9.2 to 2.9.4 ([#362](https://github.com/statnett/image-scanner-operator/issues/362)) ([d716d48](https://github.com/statnett/image-scanner-operator/commit/d716d4814289540e76f408df3283462505830c5f))
* **deps:** bump github.com/prometheus/client_golang from 1.15.0 to 1.15.1 ([#361](https://github.com/statnett/image-scanner-operator/issues/361)) ([4433b65](https://github.com/statnett/image-scanner-operator/commit/4433b6539701ebfff8dcda7756b7057e28774f4d))

## [0.5.3](https://github.com/statnett/image-scanner-operator/compare/v0.5.2...v0.5.3) (2023-05-02)


### Bug Fixes

* improve logging when CIS .status.conditions patch fails ([#357](https://github.com/statnett/image-scanner-operator/issues/357)) ([beece44](https://github.com/statnett/image-scanner-operator/commit/beece443a47e8002fcef6c8bc36ee054b87e454c))


### Dependency Updates

* **deps:** bump k8s.io/klog/v2 from 2.90.1 to 2.100.1 ([#355](https://github.com/statnett/image-scanner-operator/issues/355)) ([d486f1d](https://github.com/statnett/image-scanner-operator/commit/d486f1d7a05c13c728de8231ecdba66c73f48304))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.41.0 ([#348](https://github.com/statnett/image-scanner-operator/issues/348)) ([9aa6626](https://github.com/statnett/image-scanner-operator/commit/9aa6626d4d9b95975acd0072485ba40595e0e2ab))

## [0.5.2](https://github.com/statnett/image-scanner-operator/compare/v0.5.1...v0.5.2) (2023-04-26)


### Dependency Updates

* **deps:** bump github.com/prometheus/client_golang from 1.14.0 to 1.15.0 ([#333](https://github.com/statnett/image-scanner-operator/issues/333)) ([1a29063](https://github.com/statnett/image-scanner-operator/commit/1a2906359a7cff9036f8482a34d7ff12b2550cd5))
* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.7 to 0.1.8 ([#334](https://github.com/statnett/image-scanner-operator/issues/334)) ([72e3401](https://github.com/statnett/image-scanner-operator/commit/72e340115ef6f088a3fa5dd362da9e07433990e0))
* **deps:** bump github.com/vektra/mockery/v2 from 2.23.1 to 2.24.0 ([#324](https://github.com/statnett/image-scanner-operator/issues/324)) ([0721c07](https://github.com/statnett/image-scanner-operator/commit/0721c07d1275f394b854201f9fc569e09720cf64))
* **deps:** bump github.com/vektra/mockery/v2 from 2.24.0 to 2.26.1 ([#346](https://github.com/statnett/image-scanner-operator/issues/346)) ([d2b9b17](https://github.com/statnett/image-scanner-operator/commit/d2b9b175d98d5dc00f03ee061733057fe1828155))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.39.1 ([#316](https://github.com/statnett/image-scanner-operator/issues/316)) ([7a196a0](https://github.com/statnett/image-scanner-operator/commit/7a196a0b60c0bb15980a790164964d85b2ea14cb))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.40.0 ([#335](https://github.com/statnett/image-scanner-operator/issues/335)) ([1dfee65](https://github.com/statnett/image-scanner-operator/commit/1dfee65e2dbceb39c3ef882151a60755bcac71a2))

## [0.5.1](https://github.com/statnett/image-scanner-operator/compare/v0.5.0...v0.5.1) (2023-03-29)


### Dependency Updates

* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.6 to 0.1.7 ([#310](https://github.com/statnett/image-scanner-operator/issues/310)) ([83207df](https://github.com/statnett/image-scanner-operator/commit/83207df2c252666ab5ffe7c52a6f115254eb11cc))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.14.5 to 0.14.6 ([#311](https://github.com/statnett/image-scanner-operator/issues/311)) ([89277ed](https://github.com/statnett/image-scanner-operator/commit/89277ed4ddda3e2b0503243c80709f2e3457b0a0))

## [0.5.0](https://github.com/statnett/image-scanner-operator/compare/v0.4.8...v0.5.0) (2023-03-24)


### ⚠ BREAKING CHANGES

* run scan jobs in own namespace by default ([#256](https://github.com/statnett/image-scanner-operator/issues/256))

### Features

* run scan jobs in own namespace by default ([#256](https://github.com/statnett/image-scanner-operator/issues/256)) ([9259144](https://github.com/statnett/image-scanner-operator/commit/92591446a8cee6bf4b0fe3a03b4b0ac36a1eb8d9))


### Bug Fixes

* tighten cluster-wide RBAC ([#253](https://github.com/statnett/image-scanner-operator/issues/253)) ([2d4014f](https://github.com/statnett/image-scanner-operator/commit/2d4014f1982843014c1b0506e7010a6350d60c17))


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.9.1 to 2.9.2 ([#295](https://github.com/statnett/image-scanner-operator/issues/295)) ([4fe6739](https://github.com/statnett/image-scanner-operator/commit/4fe6739b68dd211d04489990a0061edcbacfcf0b))
* **deps:** bump github.com/onsi/gomega from 1.27.4 to 1.27.5 ([#296](https://github.com/statnett/image-scanner-operator/issues/296)) ([e9b37a7](https://github.com/statnett/image-scanner-operator/commit/e9b37a7f0eadd772259e554fa34f873fa29d435e))
* **deps:** bump github.com/vektra/mockery/v2 from 2.22.1 to 2.23.0 ([#286](https://github.com/statnett/image-scanner-operator/issues/286)) ([a53f876](https://github.com/statnett/image-scanner-operator/commit/a53f87630c449d04f83612766341965fda7019d5))
* **deps:** bump github.com/vektra/mockery/v2 from 2.23.0 to 2.23.1 ([#292](https://github.com/statnett/image-scanner-operator/issues/292)) ([576f74e](https://github.com/statnett/image-scanner-operator/commit/576f74ea6f74a6b6147f8bebd947cb378c5a1972))
* **deps:** bump k8s.io/api from 0.26.2 to 0.26.3 ([#288](https://github.com/statnett/image-scanner-operator/issues/288)) ([68b6931](https://github.com/statnett/image-scanner-operator/commit/68b6931caf65a223087249ef4b0eaf43af21e983))
* **deps:** bump k8s.io/apimachinery from 0.26.2 to 0.26.3 ([#287](https://github.com/statnett/image-scanner-operator/issues/287)) ([b9f1c57](https://github.com/statnett/image-scanner-operator/commit/b9f1c574e404bdd4653edb1d595feb1fd434f7f6))
* **deps:** bump k8s.io/client-go from 0.26.2 to 0.26.3 ([#289](https://github.com/statnett/image-scanner-operator/issues/289)) ([cd5fbd1](https://github.com/statnett/image-scanner-operator/commit/cd5fbd128fd98be7bdae7f5bddd0b102614ba1ae))

## [0.4.8](https://github.com/statnett/image-scanner-operator/compare/v0.4.7...v0.4.8) (2023-03-14)


### Dependency Updates

* **deps:** bump github.com/onsi/gomega from 1.27.2 to 1.27.4 ([#277](https://github.com/statnett/image-scanner-operator/issues/277)) ([517804f](https://github.com/statnett/image-scanner-operator/commit/517804f81525d8721a371a3803d1c2678614d887))
* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.5 to 0.1.6 ([#264](https://github.com/statnett/image-scanner-operator/issues/264)) ([0d20cc1](https://github.com/statnett/image-scanner-operator/commit/0d20cc16121d48a189526c1a96b941155b61b20c))
* **deps:** bump github.com/vektra/mockery/v2 from 2.21.4 to 2.22.1 ([#265](https://github.com/statnett/image-scanner-operator/issues/265)) ([e3b9c80](https://github.com/statnett/image-scanner-operator/commit/e3b9c80792fb8f0cd49b7174a51d5d237779b1e7))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.38.3 ([#279](https://github.com/statnett/image-scanner-operator/issues/279)) ([7f2e089](https://github.com/statnett/image-scanner-operator/commit/7f2e0892e1fcfdf8bbc1a072131c8a4426885c36))

## [0.4.7](https://github.com/statnett/image-scanner-operator/compare/v0.4.6...v0.4.7) (2023-03-09)


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.8.4 to 2.9.0 ([#258](https://github.com/statnett/image-scanner-operator/issues/258)) ([90a4a3a](https://github.com/statnett/image-scanner-operator/commit/90a4a3a90412fff8bc4585f5841bca10eeabb857))
* **deps:** bump github.com/vektra/mockery/v2 from 2.20.2 to 2.21.4 ([#262](https://github.com/statnett/image-scanner-operator/issues/262)) ([e888f71](https://github.com/statnett/image-scanner-operator/commit/e888f71f55a911eab902ab508ba934746a58928f))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.38.2 ([#261](https://github.com/statnett/image-scanner-operator/issues/261)) ([3fdbffc](https://github.com/statnett/image-scanner-operator/commit/3fdbffc3789e3301ba7e9836c1894d241cd34d76))

## [0.4.6](https://github.com/statnett/image-scanner-operator/compare/v0.4.5...v0.4.6) (2023-03-03)


### Dependency Updates

* **deps:** bump github.com/onsi/gomega from 1.27.1 to 1.27.2 ([#244](https://github.com/statnett/image-scanner-operator/issues/244)) ([a6a2b39](https://github.com/statnett/image-scanner-operator/commit/a6a2b39600fd0335cdb0cb7619f8ec578d74d5c3))
* **deps:** bump k8s.io/api from 0.26.1 to 0.26.2 ([#250](https://github.com/statnett/image-scanner-operator/issues/250)) ([e29297d](https://github.com/statnett/image-scanner-operator/commit/e29297d54f4dc00feb00b6d587b03109a506d6a6))
* **deps:** bump k8s.io/apimachinery from 0.26.1 to 0.26.2 ([#247](https://github.com/statnett/image-scanner-operator/issues/247)) ([f9dd8f1](https://github.com/statnett/image-scanner-operator/commit/f9dd8f128b91fc57093e5b87e5aae3f105182072))
* **deps:** bump k8s.io/client-go from 0.26.1 to 0.26.2 ([#248](https://github.com/statnett/image-scanner-operator/issues/248)) ([d13ade2](https://github.com/statnett/image-scanner-operator/commit/d13ade24be18e76738c8c8095e0034c86f0ff42a))
* **deps:** bump k8s.io/klog/v2 from 2.90.0 to 2.90.1 ([#251](https://github.com/statnett/image-scanner-operator/issues/251)) ([67fb92d](https://github.com/statnett/image-scanner-operator/commit/67fb92dea31a4bfa8960b0c7dd4b319116244c0e))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.14.4 to 0.14.5 ([#249](https://github.com/statnett/image-scanner-operator/issues/249)) ([69cd4e8](https://github.com/statnett/image-scanner-operator/commit/69cd4e821d20ed1ed0eef8ac8eb560a5042a0f15))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.38.0 ([#245](https://github.com/statnett/image-scanner-operator/issues/245)) ([2370648](https://github.com/statnett/image-scanner-operator/commit/23706485ae54b198bbf8f1ed72c0ded3fe5e421f))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.38.1 ([#252](https://github.com/statnett/image-scanner-operator/issues/252)) ([5bfcf45](https://github.com/statnett/image-scanner-operator/commit/5bfcf458b9284df1ed69071c9c7f9a8ae6ab4998))

## [0.4.5](https://github.com/statnett/image-scanner-operator/compare/v0.4.4...v0.4.5) (2023-02-28)


### Bug Fixes

* requeue event when no container state waiting found ([#241](https://github.com/statnett/image-scanner-operator/issues/241)) ([f8199b7](https://github.com/statnett/image-scanner-operator/commit/f8199b7868b1a845681dd48e21d470d652b86e02))


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.8.3 to 2.8.4 ([#242](https://github.com/statnett/image-scanner-operator/issues/242)) ([f216a4d](https://github.com/statnett/image-scanner-operator/commit/f216a4d2e809f2917f4410fd7d9eb66b6b06487d))

## [0.4.4](https://github.com/statnett/image-scanner-operator/compare/v0.4.3...v0.4.4) (2023-02-27)


### Bug Fixes

* add debug logging for container state waiting ([#236](https://github.com/statnett/image-scanner-operator/issues/236)) ([3ce31e8](https://github.com/statnett/image-scanner-operator/commit/3ce31e86d70ef3580e7bf1851543d0ee60a6c9d5))


### Dependency Updates

* **deps:** bump github.com/stretchr/testify from 1.8.1 to 1.8.2 ([#240](https://github.com/statnett/image-scanner-operator/issues/240)) ([5d50b3d](https://github.com/statnett/image-scanner-operator/commit/5d50b3da1b49320cb0c74ab7b9a3d097a59bd534))

## [0.4.3](https://github.com/statnett/image-scanner-operator/compare/v0.4.2...v0.4.3) (2023-02-24)


### Bug Fixes

* ensure job pod is deleted when deleting scan job ([#234](https://github.com/statnett/image-scanner-operator/issues/234)) ([380e3bd](https://github.com/statnett/image-scanner-operator/commit/380e3bd5711c7e66571a257c51dd55914a4ea712))


### Dependency Updates

* **deps:** bump github.com/vektra/mockery/v2 from 2.20.0 to 2.20.2 ([#233](https://github.com/statnett/image-scanner-operator/issues/233)) ([285e46b](https://github.com/statnett/image-scanner-operator/commit/285e46ba7c63faefaaa76a8b55a191ad50d87671))

## [0.4.2](https://github.com/statnett/image-scanner-operator/compare/v0.4.1...v0.4.2) (2023-02-23)


### Bug Fixes

* ensure back-off scan jobs are deleted ([#229](https://github.com/statnett/image-scanner-operator/issues/229)) ([9b9a8d2](https://github.com/statnett/image-scanner-operator/commit/9b9a8d2c5101406d627871c22e8f7edc8c19ac91))
* max active scan job limit ([#226](https://github.com/statnett/image-scanner-operator/issues/226)) ([6da1636](https://github.com/statnett/image-scanner-operator/commit/6da163659c787cd6842efa0bbfcc566872a17f56))
* request smaller PV for trivy server cache ([#227](https://github.com/statnett/image-scanner-operator/issues/227)) ([26ee6cb](https://github.com/statnett/image-scanner-operator/commit/26ee6cb1e80ea2daec31287fbdb0b896439d5282))
* use last scan job UID instead of name ([#217](https://github.com/statnett/image-scanner-operator/issues/217)) ([41a0a4b](https://github.com/statnett/image-scanner-operator/commit/41a0a4b3478fa0d183e77842fd36fe502e93be4c))


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.8.1 to 2.8.3 ([#222](https://github.com/statnett/image-scanner-operator/issues/222)) ([ee67368](https://github.com/statnett/image-scanner-operator/commit/ee67368bf70e6776254b1dd18c00183da7595222))
* **deps:** bump github.com/onsi/gomega from 1.27.0 to 1.27.1 ([#220](https://github.com/statnett/image-scanner-operator/issues/220)) ([ceaf6a0](https://github.com/statnett/image-scanner-operator/commit/ceaf6a05f453d3c731783f7c38005c5b718ec347))
* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.4 to 0.1.5 ([#221](https://github.com/statnett/image-scanner-operator/issues/221)) ([605164f](https://github.com/statnett/image-scanner-operator/commit/605164f4a8ee49e28a9b898ff75d2f21705646a2))

## [0.4.1](https://github.com/statnett/image-scanner-operator/compare/v0.4.0...v0.4.1) (2023-02-17)


### Bug Fixes

* add missing RBAC and ensure job pods are deleted ([#216](https://github.com/statnett/image-scanner-operator/issues/216)) ([fb29ba2](https://github.com/statnett/image-scanner-operator/commit/fb29ba2d04d07f74ca27c28e0c3525705db11dfe))
* rescan machinery ([#213](https://github.com/statnett/image-scanner-operator/issues/213)) ([858f4a2](https://github.com/statnett/image-scanner-operator/commit/858f4a2c4f67d13a8b0f1cc712bef2e77fb58f9a))


### Dependency Updates

* **deps:** bump github.com/onsi/gomega from 1.26.0 to 1.27.0 ([#214](https://github.com/statnett/image-scanner-operator/issues/214)) ([5b4d88a](https://github.com/statnett/image-scanner-operator/commit/5b4d88acce18bd2f8c0508e640ae9aab13244819))

## [0.4.0](https://github.com/statnett/image-scanner-operator/compare/v0.3.1...v0.4.0) (2023-02-16)


### ⚠ BREAKING CHANGES

* introduce mitchellh/hashstructure for hashing ([#212](https://github.com/statnett/image-scanner-operator/issues/212))

### Bug Fixes

* introduce mitchellh/hashstructure for hashing ([#212](https://github.com/statnett/image-scanner-operator/issues/212)) ([26eefd7](https://github.com/statnett/image-scanner-operator/commit/26eefd72fe623ba705e46035643b721cf6b64878))
* rescan based on ticker ([#208](https://github.com/statnett/image-scanner-operator/issues/208)) ([bc87fcd](https://github.com/statnett/image-scanner-operator/commit/bc87fcdca5fff2ce35b91264c276bf4c430e2636))


### Dependency Updates

* **deps:** bump github.com/onsi/ginkgo/v2 from 2.8.0 to 2.8.1 ([#209](https://github.com/statnett/image-scanner-operator/issues/209)) ([258b0d3](https://github.com/statnett/image-scanner-operator/commit/258b0d380d82e75280ec0d3121c8d50dab2552fb))
* **deps:** bump github.com/vektra/mockery/v2 from 2.18.0 to 2.20.0 ([#206](https://github.com/statnett/image-scanner-operator/issues/206)) ([a94425f](https://github.com/statnett/image-scanner-operator/commit/a94425fca6dbeb315cb673b9b6242456765c55e3))
* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.37.3 ([#211](https://github.com/statnett/image-scanner-operator/issues/211)) ([82d54d5](https://github.com/statnett/image-scanner-operator/commit/82d54d588edfe92200930ff0d3abc689512d0b53))

## [0.3.1](https://github.com/statnett/image-scanner-operator/compare/v0.3.0...v0.3.1) (2023-02-10)


### Bug Fixes

* always set RequeueAfter when reconciling CIS ([#198](https://github.com/statnett/image-scanner-operator/issues/198)) ([ad8d921](https://github.com/statnett/image-scanner-operator/commit/ad8d921a87f098917c6d2b04e7e22184428fa0be))


### Dependency Updates

* **deps:** update ghcr.io/aquasecurity/trivy docker tag to v0.37.2 ([#199](https://github.com/statnett/image-scanner-operator/issues/199)) ([f330486](https://github.com/statnett/image-scanner-operator/commit/f33048618a1e5efc7e8606779dd069331e1e544c))

## [0.3.0](https://github.com/statnett/image-scanner-operator/compare/v0.2.2...v0.3.0) (2023-02-09)


### ⚠ BREAKING CHANGES

* should have unique config map names ([#163](https://github.com/statnett/image-scanner-operator/issues/163))

### Bug Fixes

* never retry scan job ([#194](https://github.com/statnett/image-scanner-operator/issues/194)) ([0198fc1](https://github.com/statnett/image-scanner-operator/commit/0198fc1da0b64ccbf327d4e739cbd3c8fd23648a))
* should have unique config map names ([#163](https://github.com/statnett/image-scanner-operator/issues/163)) ([b7a8aa3](https://github.com/statnett/image-scanner-operator/commit/b7a8aa3fb805ac07c7906b0dde6c877c08a25d08))


### Dependency Updates

* **deps:** bump github.com/statnett/controller-runtime-viper from 0.1.3 to 0.1.4 ([#190](https://github.com/statnett/image-scanner-operator/issues/190)) ([b2e152f](https://github.com/statnett/image-scanner-operator/commit/b2e152f0e3c4e956b617ab85a44ab97381959e77))
* **deps:** bump github.com/vektra/mockery/v2 from 2.16.0 to 2.18.0 ([#188](https://github.com/statnett/image-scanner-operator/issues/188)) ([7aa53e2](https://github.com/statnett/image-scanner-operator/commit/7aa53e2083a1619ab95e43c5d50340fab5282583))
* **deps:** bump sigs.k8s.io/controller-runtime from 0.14.2 to 0.14.4 ([#187](https://github.com/statnett/image-scanner-operator/issues/187)) ([03fcc85](https://github.com/statnett/image-scanner-operator/commit/03fcc854e2c103777a44e3f257fce38db610ad06))

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
