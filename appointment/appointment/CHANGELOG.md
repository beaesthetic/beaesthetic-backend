# Changelog

## [1.9.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-1.8.0...appointment-service-v1.9.0) (2025-12-13)


### Features

* **agenda:** added recovery stuck event as startup task ([59307a3](https://github.com/beaesthetic/beaesthetic-backend/commit/59307a3f77205bdd60299af3a75580232b5fe9d1))
* **agenda:** enable all customer notifications ([#7](https://github.com/beaesthetic/beaesthetic-backend/issues/7)) ([0da5314](https://github.com/beaesthetic/beaesthetic-backend/commit/0da5314b3de33ce4d962b373297d0e5f9641c12a))
* **agenda:** notification on creation/reschedule with whitelist ([#5](https://github.com/beaesthetic/beaesthetic-backend/issues/5)) ([663e219](https://github.com/beaesthetic/beaesthetic-backend/commit/663e219e9fd1b2a7fff741b9f4a80b3d4a9e9363))
* **agenda:** reminder failed monitor ([0c6c943](https://github.com/beaesthetic/beaesthetic-backend/commit/0c6c94355f3964ec2c0c0911bbabf0d1732893c5))
* **agenda:** tracing reminder metrics + kotlin2 ([3b3ab07](https://github.com/beaesthetic/beaesthetic-backend/commit/3b3ab07136a57799f712b5dfe06c8177af4e8226))
* **appointment:** added scheduler and notification integration ([71e6b7f](https://github.com/beaesthetic/beaesthetic-backend/commit/71e6b7f8acdadf3ec72226318b6dbfc60ccac23e))
* **appointment:** added scheduler and notification integration ([8bc7708](https://github.com/beaesthetic/beaesthetic-backend/commit/8bc770862b1c6178f6baf9132843345fce0c38c7))
* **appointment:** added services domain ([19a9bb9](https://github.com/beaesthetic/beaesthetic-backend/commit/19a9bb9698e34c8d11d99ad9cd2c82dbfac6712e))
* **core:** adding appointment service ([#3](https://github.com/beaesthetic/beaesthetic-backend/issues/3)) ([ec02d53](https://github.com/beaesthetic/beaesthetic-backend/commit/ec02d536e8ee07d4ed4e23f700a6654b8fba1671))


### Bug Fixes

* **agenda:** add missing locales ([1c8c23f](https://github.com/beaesthetic/beaesthetic-backend/commit/1c8c23fe64af002c2e1c87e8245083d3391ee724))
* **agenda:** added threshold for reminders ([397102c](https://github.com/beaesthetic/beaesthetic-backend/commit/397102cd2704ee74a8e0e7a69062ba05f50231c8))
* **agenda:** added threshold proper threshold and reminder time on helm ([63416a6](https://github.com/beaesthetic/beaesthetic-backend/commit/63416a66a2ef453ea1f26715a2f1e24aae40a5ae))
* **agenda:** deserialization pending notification ([03902d6](https://github.com/beaesthetic/beaesthetic-backend/commit/03902d640cc3becffb7cb8643ba65441821bcc69))
* **agenda:** detect ([0957970](https://github.com/beaesthetic/beaesthetic-backend/commit/0957970cdff83029e8ab4f14842adcc0171fb611))
* **agenda:** filter out cancelled event for customer ([c7c4e69](https://github.com/beaesthetic/beaesthetic-backend/commit/c7c4e696e571e71c410a8a53196055db75a9fe3a))
* **agenda:** otel grpc endpoint ([7148c52](https://github.com/beaesthetic/beaesthetic-backend/commit/7148c52e8b0d8cb919b1d84a56749c8c344f601d))
* **agenda:** pending notification reflection native handle ([2c629f8](https://github.com/beaesthetic/beaesthetic-backend/commit/2c629f8a36ef432e275c3d22bcf182d68d3312ff))
* **agenda:** register reminder times up ([a77d02d](https://github.com/beaesthetic/beaesthetic-backend/commit/a77d02dcf46fee9908a36cac58d15d3251cb6093))
* **agenda:** reminder failed monitor ([846ea1e](https://github.com/beaesthetic/beaesthetic-backend/commit/846ea1edb8c8806d1b424601e3f72c5072d97afb))
* **agenda:** trying register json object for reflection ([9099dcb](https://github.com/beaesthetic/beaesthetic-backend/commit/9099dcb3ace940ea488bcf529798e5ee38f8b8e4))
* **agenda:** update agenda entity name ([22308ff](https://github.com/beaesthetic/beaesthetic-backend/commit/22308ff8cfc897e88050247d7d247e948bbddf1d))
* **agenda:** update query find by attendee id ([8400f16](https://github.com/beaesthetic/beaesthetic-backend/commit/8400f169f12aea7f7eb00323419d9cb90522ce97))
* appointment admin api ([9050c65](https://github.com/beaesthetic/beaesthetic-backend/commit/9050c6572ce4ae36989412e90cfd02c6699f567c))
* appointment admin api ([e27afdd](https://github.com/beaesthetic/beaesthetic-backend/commit/e27afdd7fec6506de0a201bb6ad4bb8cbcad97ae))
* **appointment:** added appoinmtment to releaserc ([0c303d1](https://github.com/beaesthetic/beaesthetic-backend/commit/0c303d197da6f76c9f1ea046b2aa5357ed9d2e2b))
* **appointment:** filter out cancelled appoitments ([68500de](https://github.com/beaesthetic/beaesthetic-backend/commit/68500de5de1797884389d3f72f23018a1f0175dd))
* **appointment:** invalidate services cache ([2fe4584](https://github.com/beaesthetic/beaesthetic-backend/commit/2fe45848947bd8c11c61ca3d9ed1476ce1da86e2))
* **appointment:** missing reflection registration ([51795e2](https://github.com/beaesthetic/beaesthetic-backend/commit/51795e22fd5ee44e4b9c2c3b3a4a14370ff93dae))
* **appointment:** remove composite kotlin build + refactor docker ([29444cc](https://github.com/beaesthetic/beaesthetic-backend/commit/29444ccaec341554a95b651fda784f78c8ee120b))
* **appointment:** remove composite kotlin build + refactor docker ([d5ae28a](https://github.com/beaesthetic/beaesthetic-backend/commit/d5ae28ae18e07e68cab1e3570ae95be7ab9a78e6))
* **appointment:** remove composite kotlin build + refactor docker ([f3f1fa6](https://github.com/beaesthetic/beaesthetic-backend/commit/f3f1fa62221c77213c47739e50451785604a5b3b))
* **appointment:** remove confirm queue declaration ([596fb64](https://github.com/beaesthetic/beaesthetic-backend/commit/596fb642ca8ba6e6d0ee85b87126d8bd50a1e743))
* **appointment:** wrong route specified to scheduler ([0e0f291](https://github.com/beaesthetic/beaesthetic-backend/commit/0e0f2913779d52c26356137366504f2f363b287c))
* **apppointment:** working mongo deserialization ([85e6d1a](https://github.com/beaesthetic/beaesthetic-backend/commit/85e6d1a8d0c256856d98af3ded2617a45cc15e0a))
* **apppointment:** working mongo deserialization ([eeb58c6](https://github.com/beaesthetic/beaesthetic-backend/commit/eeb58c6540c514a3c2a6fd40b35aa5b36788eb0f))
* **deps:** update appointment service ([#22](https://github.com/beaesthetic/beaesthetic-backend/issues/22)) ([448f2ee](https://github.com/beaesthetic/beaesthetic-backend/commit/448f2ee91255bc2e176797e208a0b6ae4d564c86))
* gradle build files ([b227dd9](https://github.com/beaesthetic/beaesthetic-backend/commit/b227dd9d343f1dbbdfa8f35bca9d77cf7a458556))
