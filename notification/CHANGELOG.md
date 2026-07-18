# Changelog

## [2.6.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.5.0...notification-service-v2.6.0) (2026-07-18)


### Features

* remove legacy notification domain and related infrastructure ([ef4086a](https://github.com/beaesthetic/beaesthetic-backend/commit/ef4086a2fe1ba1d9c62e88acb6f8fe40e79a1cca))

## [2.5.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.4.0...notification-service-v2.5.0) (2026-07-18)


### Features

* implement customer notification service with SMS support ([5a7d9b2](https://github.com/beaesthetic/beaesthetic-backend/commit/5a7d9b20ae9e72804889ca1de4c3b94434ce4e3c))
* implement protobuf-based customer notification system with AMQP integration ([d37a1b0](https://github.com/beaesthetic/beaesthetic-backend/commit/d37a1b07838113ae3979b2dffa79a135577a835c))
* notification revamp ([#86](https://github.com/beaesthetic/beaesthetic-backend/issues/86)) ([c9249a6](https://github.com/beaesthetic/beaesthetic-backend/commit/c9249a62602357eb0cd7edfce3f662f3b2070ce7))
* update release notification workflow and helm values for notification service ([7f612c9](https://github.com/beaesthetic/beaesthetic-backend/commit/7f612c93c74082be2b6d29617f6a62f65803ac52))
* update release notification workflow, enhance ApplicationSet for notification service, and set image tag in Helm values ([b949d50](https://github.com/beaesthetic/beaesthetic-backend/commit/b949d50fe0934854921fca0d17abf970f19ee34d))

## [2.4.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.3.0...notification-service-v2.4.0) (2026-07-09)


### Features

* enhance notification error handling and add unit test for SMS gateway notification ([ef28dd7](https://github.com/beaesthetic/beaesthetic-backend/commit/ef28dd74c9610ad6035437d587e2fb3fc2c182d9))
* refactor database name generation in helm templates for appointment, notification, and scheduler ([e952752](https://github.com/beaesthetic/beaesthetic-backend/commit/e952752b1f73d4d60522e57e2da64ac82197b2b1))
* set default environment for rabbitmqVhost in helm templates ([277c8ce](https://github.com/beaesthetic/beaesthetic-backend/commit/277c8cee2e68524a2abf6eaae1f6b6b49e485a76))
* update external secret configuration for SMS gateway API key ([58573c2](https://github.com/beaesthetic/beaesthetic-backend/commit/58573c2e4df377c68fe7385491112078248301f7))

## [2.3.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.2.0...notification-service-v2.3.0) (2026-07-07)


### Features

* add Helm templates and configuration for appointment, notification, and scheduler services ([121f6d2](https://github.com/beaesthetic/beaesthetic-backend/commit/121f6d2b8faa7e97cf10caa92570710213003b1b))
* add Helm values files for appointment, customer, notification, and scheduler services ([611f314](https://github.com/beaesthetic/beaesthetic-backend/commit/611f31424ffe45ac75de9028cf1da9450fb54958))
* add RabbitMQ vhost helpers and update configurations for appointment, notification, and scheduler services ([a2d8f6d](https://github.com/beaesthetic/beaesthetic-backend/commit/a2d8f6d7442c45be20315491438106fe0b846acb))
* add renderEnvConfig helper for appointment, notification, and scheduler services ([cb048a7](https://github.com/beaesthetic/beaesthetic-backend/commit/cb048a75e659c8b7c116075ae10b096057339e36))
* refactor Helm templates to use common labels and environment configurations for appointment, notification, and scheduler services ([5f8d8ab](https://github.com/beaesthetic/beaesthetic-backend/commit/5f8d8abff7de64687ac30c653ccf31839bd2efe6))
* update database name formatting and add common labels for appointment, notification, and scheduler services ([1b1646e](https://github.com/beaesthetic/beaesthetic-backend/commit/1b1646e431d23b2e58f6b683cc95cb9288374a2b))


### Bug Fixes

* ensure file source is imported for database migrations ([0f8a9a8](https://github.com/beaesthetic/beaesthetic-backend/commit/0f8a9a8abc74a169726515e8c1c0babb57f6f6a4))
* update build command to include all cmd subdirectories ([178542c](https://github.com/beaesthetic/beaesthetic-backend/commit/178542c6a5b7ee61c0857c4bf82df7ad17f8b4af))

## [2.2.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.1.2...notification-service-v2.2.0) (2026-07-04)


### Features

* add health check endpoint and SMS webhook handling ([2a11f5e](https://github.com/beaesthetic/beaesthetic-backend/commit/2a11f5e2cea086ff73919abfd47b5172117c94fc))
* add OpenAPI specifications for notification and SMS webhook APIs ([ae279cf](https://github.com/beaesthetic/beaesthetic-backend/commit/ae279cf4f37fa74c602bbfeddec67a06157694fb))
* implement dependency injection for notification service and HTTP server ([2a11f5e](https://github.com/beaesthetic/beaesthetic-backend/commit/2a11f5e2cea086ff73919abfd47b5172117c94fc))
* refactor HTTP server initialization to use net/http package ([6fb425e](https://github.com/beaesthetic/beaesthetic-backend/commit/6fb425ea7db4442501cb3d8a1968568acd10ab6f))


### Code Refactoring

* restructure notification service and remove deprecated container code ([2a11f5e](https://github.com/beaesthetic/beaesthetic-backend/commit/2a11f5e2cea086ff73919abfd47b5172117c94fc))

## [2.1.2](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.1.1...notification-service-v2.1.2) (2026-07-02)


### Code Refactoring

* remove backfill command and related dependencies ([4fd7001](https://github.com/beaesthetic/beaesthetic-backend/commit/4fd70019fa44ded62c7177e7bb3fed16ef19e12c))

## [2.1.1](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.1.0...notification-service-v2.1.1) (2026-07-02)


### Bug Fixes

* enable schema initialization in outbox configuration ([1509358](https://github.com/beaesthetic/beaesthetic-backend/commit/15093587fd765c5d36e01fe0cf6be66845881e46))
* update NewNotification to require only id and content, allow empty title ([c6e4347](https://github.com/beaesthetic/beaesthetic-backend/commit/c6e43471c223f154a11200576b2874e6f24a9453))

## [2.1.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v2.0.0...notification-service-v2.1.0) (2026-07-02)


### Features

* refactor notification service and integrate outbox pattern with RabbitMQ ([eef38cd](https://github.com/beaesthetic/beaesthetic-backend/commit/eef38cd5d3aadc116c1fee2360583e62d4a84dc2))

## [2.0.0](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v1.1.2...notification-service-v2.0.0) (2026-07-02)


### ⚠ BREAKING CHANGES

* **notification:** migrate notification service to Go

### Features

* **notification:** migrate notification service to Go ([d7b2387](https://github.com/beaesthetic/beaesthetic-backend/commit/d7b238786623c1bdc05f85706e442a7fa414e0eb))

## [1.1.2](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v1.1.1...notification-service-v1.1.2) (2025-12-16)


### Bug Fixes

* docker image base ([860a009](https://github.com/beaesthetic/beaesthetic-backend/commit/860a0090b2f1fc20fb07626cef234a8e217beff6))

## [1.1.1](https://github.com/beaesthetic/beaesthetic-backend/compare/notification-service-v1.1.0...notification-service-v1.1.1) (2025-12-15)


### Bug Fixes

* **deps:** update notification service ([#36](https://github.com/beaesthetic/beaesthetic-backend/issues/36)) ([149fc6b](https://github.com/beaesthetic/beaesthetic-backend/commit/149fc6b69d501aa0033d231c1b791ce99e48f95c))
* gradle improve customers ([77da985](https://github.com/beaesthetic/beaesthetic-backend/commit/77da9850a6e3556d33374f89c9de272fad3007b1))
* gradle improve notification ([9590245](https://github.com/beaesthetic/beaesthetic-backend/commit/9590245481640e445b0f714740e485a39927637c))
