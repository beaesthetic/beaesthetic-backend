# Changelog

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
