# Changelog

## [1.5.0](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.4.0...scheduler-service-v1.5.0) (2026-07-09)


### Features

* refactor database name generation in helm templates for appointment, notification, and scheduler ([e952752](https://github.com/beaesthetic/beaesthetic-backend/commit/e952752b1f73d4d60522e57e2da64ac82197b2b1))
* set default environment for rabbitmqVhost in helm templates ([277c8ce](https://github.com/beaesthetic/beaesthetic-backend/commit/277c8cee2e68524a2abf6eaae1f6b6b49e485a76))

## [1.4.0](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.3.1...scheduler-service-v1.4.0) (2026-07-07)


### Features

* add Helm templates and configuration for appointment, notification, and scheduler services ([121f6d2](https://github.com/beaesthetic/beaesthetic-backend/commit/121f6d2b8faa7e97cf10caa92570710213003b1b))
* add Helm values files for appointment, customer, notification, and scheduler services ([611f314](https://github.com/beaesthetic/beaesthetic-backend/commit/611f31424ffe45ac75de9028cf1da9450fb54958))
* add RabbitMQ vhost helpers and update configurations for appointment, notification, and scheduler services ([a2d8f6d](https://github.com/beaesthetic/beaesthetic-backend/commit/a2d8f6d7442c45be20315491438106fe0b846acb))
* add renderEnvConfig helper for appointment, notification, and scheduler services ([cb048a7](https://github.com/beaesthetic/beaesthetic-backend/commit/cb048a75e659c8b7c116075ae10b096057339e36))
* refactor Helm templates to use common labels and environment configurations for appointment, notification, and scheduler services ([5f8d8ab](https://github.com/beaesthetic/beaesthetic-backend/commit/5f8d8abff7de64687ac30c653ccf31839bd2efe6))
* update database name formatting and add common labels for appointment, notification, and scheduler services ([1b1646e](https://github.com/beaesthetic/beaesthetic-backend/commit/1b1646e431d23b2e58f6b683cc95cb9288374a2b))

## [1.3.1](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.3.0...scheduler-service-v1.3.1) (2026-07-02)


### Bug Fixes

* update RabbitMQ configuration in values.yaml ([2b089e3](https://github.com/beaesthetic/beaesthetic-backend/commit/2b089e3db21ff4bc952217f6bb9b16ab3dbb6f7f))

## [1.3.0](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.2.0...scheduler-service-v1.3.0) (2026-07-02)


### Features

* add external secrets for PostgreSQL and RabbitMQ configuration ([9af1c75](https://github.com/beaesthetic/beaesthetic-backend/commit/9af1c75ab89514655c8c261347662cb5599c1281))
* add golangci-lint configuration and update magefile for linting tools ([3072e99](https://github.com/beaesthetic/beaesthetic-backend/commit/3072e9932de3d7e3f01732b4270acb4de5d7d0f4))


### Bug Fixes

* update Dockerfile and workflow to use ubuntu-latest and improve build process ([fc3946a](https://github.com/beaesthetic/beaesthetic-backend/commit/fc3946a2918e84198d678633edbda93f620de96c))
* update postgres-dsn in external secret to point to the correct database ([7aadb73](https://github.com/beaesthetic/beaesthetic-backend/commit/7aadb734c20be402b86bececf9705429abbd74bf))
* update RabbitMQ configuration in values.yaml for correct credentials and host ([1a388ea](https://github.com/beaesthetic/beaesthetic-backend/commit/1a388ea6d792d2dd41b16234bcd112062584876f))


### Code Refactoring

* remove legacy Redis migration support and related configurations ([448277f](https://github.com/beaesthetic/beaesthetic-backend/commit/448277f017997ff51e45f809ea3abf45088c3b7f))

## [1.2.0](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.1.8...scheduler-service-v1.2.0) (2026-06-29)


### Features

* implement OpenAPI server handlers for scheduling operations ([311df48](https://github.com/beaesthetic/beaesthetic-backend/commit/311df4873b28e81f3c47a09543923d4f637378f9))

## [1.1.8](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.1.7...scheduler-service-v1.1.8) (2025-12-23)


### Bug Fixes

* scheduler base docker image ([b25dea5](https://github.com/beaesthetic/beaesthetic-backend/commit/b25dea56bac97764cb1d9bba28a1a450c84b5400))

## [1.1.7](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.1.6...scheduler-service-v1.1.7) (2025-12-16)


### Bug Fixes

* docker image base ([3efc336](https://github.com/beaesthetic/beaesthetic-backend/commit/3efc336c8d2e242810db022b21d3e0a2b3447da3))

## [1.1.6](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.1.5...scheduler-service-v1.1.6) (2025-12-16)


### Bug Fixes

* docker image base ([860a009](https://github.com/beaesthetic/beaesthetic-backend/commit/860a0090b2f1fc20fb07626cef234a8e217beff6))

## [1.1.5](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.1.4...scheduler-service-v1.1.5) (2025-12-15)


### Bug Fixes

* added some appointment tests + gradle improvement ([6247e9e](https://github.com/beaesthetic/beaesthetic-backend/commit/6247e9edcb319669371ae5774f1f4f1b65982d62))
* gradle improve scheduler ([f70c404](https://github.com/beaesthetic/beaesthetic-backend/commit/f70c40495a88281c9ad884526bfce52981290376))

## [1.1.4](https://github.com/beaesthetic/beaesthetic-backend/compare/scheduler-service-v1.1.3...scheduler-service-v1.1.4) (2025-12-13)


### Bug Fixes

* **deps:** update scheduler service ([#16](https://github.com/beaesthetic/beaesthetic-backend/issues/16)) ([486392c](https://github.com/beaesthetic/beaesthetic-backend/commit/486392c092c85a1a75e608cd3b5007a7841a5663))
