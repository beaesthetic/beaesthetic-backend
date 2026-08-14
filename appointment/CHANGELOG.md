# Changelog

## [1.15.1](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.15.0...appointment-service-v1.15.1) (2026-08-14)


### Bug Fixes

* typo ([adfc62e](https://github.com/beaesthetic/beaesthetic-backend/commit/adfc62eaca33700edfdbd8c30ecc88798b345f9c))

## [1.15.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.14.0...appointment-service-v1.15.0) (2026-07-23)


### Features

* add reminder scheduler provider configuration to environment settings ([8af56f1](https://github.com/beaesthetic/beaesthetic-backend/commit/8af56f1f30c82932acb8bc9e522f7a4dda26b154))
* implement reminder scheduling and sending functionality with qriver ([7d68a3d](https://github.com/beaesthetic/beaesthetic-backend/commit/7d68a3debe5736230724199d2eaa2b4305f769ed))
* implement runtime process management with new runner and adapters ([23c7861](https://github.com/beaesthetic/beaesthetic-backend/commit/23c7861f94a53062770f5211431425369a5eceeb))

## [1.14.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.13.0...appointment-service-v1.14.0) (2026-07-22)


### Features

* remove backfill command and related infrastructure ([b251e63](https://github.com/beaesthetic/beaesthetic-backend/commit/b251e6352580cda7210f16c8cd180a5f26b533c8))

## [1.13.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.12.0...appointment-service-v1.13.0) (2026-07-18)


### Features

* add GitHub workflows for branches and main, remove obsolete release workflows, and update environment configurations ([892579d](https://github.com/beaesthetic/beaesthetic-backend/commit/892579dd7511802e46cacb1396599671886e664e))
* notification revamp ([#86](https://github.com/beaesthetic/beaesthetic-backend/issues/86)) ([c9249a6](https://github.com/beaesthetic/beaesthetic-backend/commit/c9249a62602357eb0cd7edfce3f662f3b2070ce7))

## [1.12.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.11.0...appointment-service-v1.12.0) (2026-07-09)


### Features

* add appointment and event reminder status functions with corresponding tests ([752c937](https://github.com/beaesthetic/beaesthetic-backend/commit/752c9379dc5ed6ae44dadf93cf5efeb040f7cea8))
* add command to schedule future appointment reminders and implement related repository methods ([de31306](https://github.com/beaesthetic/beaesthetic-backend/commit/de3130660608e2a2454eb6e1206632ef5b6c9ce3))
* add search_text column generation and related function for appointment_services ([720c389](https://github.com/beaesthetic/beaesthetic-backend/commit/720c389b9c2a8cad8ed9959b92e179b20853fd12))
* enhance logging in appointment lifecycle and messaging components ([7134861](https://github.com/beaesthetic/beaesthetic-backend/commit/7134861e20f0319190267958250387fb2d78f6a7))
* refactor database name generation in helm templates for appointment, notification, and scheduler ([e952752](https://github.com/beaesthetic/beaesthetic-backend/commit/e952752b1f73d4d60522e57e2da64ac82197b2b1))
* set default environment for rabbitmqVhost in helm templates ([277c8ce](https://github.com/beaesthetic/beaesthetic-backend/commit/277c8cee2e68524a2abf6eaae1f6b6b49e485a76))
* update MongoDB database name to remove environment suffix in values-dev.yaml ([ca15265](https://github.com/beaesthetic/beaesthetic-backend/commit/ca1526589a9cea90b90d9a5e93227c36af625585))
* update reminder status handling and enhance related tests for appointments and events ([7e07790](https://github.com/beaesthetic/beaesthetic-backend/commit/7e07790ede47829d95f7f5b9aee0afcdd4fb42b4))


### Bug Fixes

* correct formatting of namespace and labels in outbox-configmap.yaml ([2662eca](https://github.com/beaesthetic/beaesthetic-backend/commit/2662ecafd01183466dab74bbeb78d3ae1acdc97c))
* correct namespace and labels formatting in outbox-configmap.yaml ([4d6726c](https://github.com/beaesthetic/beaesthetic-backend/commit/4d6726cc394d37f77e07322879d88c454d325431))
* remove quotes from exchange name in outbox-configmap.yaml ([a4b8c9f](https://github.com/beaesthetic/beaesthetic-backend/commit/a4b8c9fb543163964ec9354491327850c5b2f68d))
* update appointment_services table schema and indexing for search optimization ([cbb6654](https://github.com/beaesthetic/beaesthetic-backend/commit/cbb665483409e6d22dedff82e056735acd169573))
* update routing_key format in outbox-configmap.yaml ([61d3fe2](https://github.com/beaesthetic/beaesthetic-backend/commit/61d3fe2e98849316ccef0fc25a263ff2526ef3e0))

## [1.11.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.10.2...appointment-service-v1.11.0) (2026-07-07)


### Features

* add Helm templates and configuration for appointment, notification, and scheduler services ([121f6d2](https://github.com/beaesthetic/beaesthetic-backend/commit/121f6d2b8faa7e97cf10caa92570710213003b1b))
* add Helm values files for appointment, customer, notification, and scheduler services ([611f314](https://github.com/beaesthetic/beaesthetic-backend/commit/611f31424ffe45ac75de9028cf1da9450fb54958))
* add RabbitMQ vhost helpers and update configurations for appointment, notification, and scheduler services ([a2d8f6d](https://github.com/beaesthetic/beaesthetic-backend/commit/a2d8f6d7442c45be20315491438106fe0b846acb))
* add renderEnvConfig helper for appointment, notification, and scheduler services ([cb048a7](https://github.com/beaesthetic/beaesthetic-backend/commit/cb048a75e659c8b7c116075ae10b096057339e36))
* refactor Helm templates to use common labels and environment configurations for appointment, notification, and scheduler services ([5f8d8ab](https://github.com/beaesthetic/beaesthetic-backend/commit/5f8d8abff7de64687ac30c653ccf31839bd2efe6))
* update database name formatting and add common labels for appointment, notification, and scheduler services ([1b1646e](https://github.com/beaesthetic/beaesthetic-backend/commit/1b1646e431d23b2e58f6b683cc95cb9288374a2b))
* update refresh policy to Periodic for external secrets in appointment service ([f93a7e5](https://github.com/beaesthetic/beaesthetic-backend/commit/f93a7e50cf993b168feb2d4cd7084e74de12e454))


### Bug Fixes

* add ENV_HTTP_ADDR to envConfig in values.yaml ([49328ad](https://github.com/beaesthetic/beaesthetic-backend/commit/49328ade09a42b0e5d9b72f299c0711611e49e8c))
* enable postgres database migration by removing underscore import ([ef9d791](https://github.com/beaesthetic/beaesthetic-backend/commit/ef9d791453b3fc553c87cf1688ac994942b9e132))
* enhance error handling and logging in various components ([6a1c8e5](https://github.com/beaesthetic/beaesthetic-backend/commit/6a1c8e5a5f5acb73aa85a238dd59eccdee92e724))
* update readiness and liveness probe paths to /health ([e4fe078](https://github.com/beaesthetic/beaesthetic-backend/commit/e4fe078039150a0a96c13de683f52dec4dd257aa))

## [1.10.2](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.10.1...appointment-service-v1.10.2) (2026-07-05)


### Bug Fixes

* correct namespace and labels formatting in Helm templates ([d901560](https://github.com/beaesthetic/beaesthetic-backend/commit/d9015600edf89ecf2d7672e1248e174ccdae722a))
* ijecting onyl new secrets ([7b6b840](https://github.com/beaesthetic/beaesthetic-backend/commit/7b6b840f9768bd9e05a8557a2c6116c6b63f4856))
* migrator missing import ([5fb947d](https://github.com/beaesthetic/beaesthetic-backend/commit/5fb947df3aaa95df10b762a9e617b9e8ab2afdbc))

## [1.10.1](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.10.0...appointment-service-v1.10.1) (2026-07-04)


### Bug Fixes

* delete appointment java ([b0dd160](https://github.com/beaesthetic/beaesthetic-backend/commit/b0dd160829e36196771e082d6b3f9963d1b45cf9))

## [1.10.0](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.9.5...appointment-service-v1.10.0) (2026-07-02)


### Features

* add RabbitMQ external secret configuration and update application properties ([19c43fc](https://github.com/beaesthetic/beaesthetic-backend/commit/19c43fcafc25b410a1db84d1c4072f85e15b5da8))

## [1.9.5](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.9.4...appointment-service-v1.9.5) (2025-12-23)


### Bug Fixes

* docker file appointment to java 21 ([a0d9a94](https://github.com/beaesthetic/beaesthetic-backend/commit/a0d9a94e3dcdb43fb62c153266e228d5d052c4d4))

## [1.9.4](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.9.3...appointment-service-v1.9.4) (2025-12-23)


### Bug Fixes

* appointment transitive dep bouncy castle native ([55a242e](https://github.com/beaesthetic/beaesthetic-backend/commit/55a242e72c603b7209b3fb1647557f2a9e942f04))

## [1.9.3](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.9.2...appointment-service-v1.9.3) (2025-12-16)


### Bug Fixes

* docker image base ([3efc336](https://github.com/beaesthetic/beaesthetic-backend/commit/3efc336c8d2e242810db022b21d3e0a2b3447da3))

## [1.9.2](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.9.1...appointment-service-v1.9.2) (2025-12-16)


### Bug Fixes

* docker image base ([860a009](https://github.com/beaesthetic/beaesthetic-backend/commit/860a0090b2f1fc20fb07626cef234a8e217beff6))

## [1.9.1](https://github.com/beaesthetic/beaesthetic-backend/compare/appointment-service-v1.9.0...appointment-service-v1.9.1) (2025-12-15)


### Bug Fixes

* added some appointment tests + gradle improvement ([6247e9e](https://github.com/beaesthetic/beaesthetic-backend/commit/6247e9edcb319669371ae5774f1f4f1b65982d62))
