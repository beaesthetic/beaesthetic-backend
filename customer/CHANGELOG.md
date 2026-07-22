# Changelog

## [1.15.0](https://github.com/beaesthetic/beaesthetic-backend/compare/customer-service-v1.14.0...customer-service-v1.15.0) (2026-07-22)


### Features

* add wallet operations and credit lots migration scripts ([6f0d785](https://github.com/beaesthetic/beaesthetic-backend/commit/6f0d7852de00b8c3756c65cf057cea6155b05d36))

## [1.14.0](https://github.com/beaesthetic/beaesthetic-backend/compare/customer-service-v1.13.0...customer-service-v1.14.0) (2026-07-22)


### Features

* add PostgreSQL queries and models for customers, fidelity cards, and wallets ([2457883](https://github.com/beaesthetic/beaesthetic-backend/commit/2457883a40d60bd3cd775baae793a38c1623f781))

## [1.13.0](https://github.com/beaesthetic/beaesthetic-backend/compare/customer-service-v1.12.0...customer-service-v1.13.0) (2026-07-11)


### Features

* add discriminator type to wallet events and implement related tests ([2164093](https://github.com/beaesthetic/beaesthetic-backend/commit/2164093696c16a17f2a7a86c1fadd89b7a76aebd))
* add redis URI to external secrets and update values.yaml for secret reference ([3f3cc58](https://github.com/beaesthetic/beaesthetic-backend/commit/3f3cc5823a25da04e889aa7f13a1a416d2a77f80))
* normalize trailing slash handling and add route aliases for fidelity and wallet APIs ([59ee46d](https://github.com/beaesthetic/beaesthetic-backend/commit/59ee46d07c6ca67e390ab21ee3a5f5d345f1e479))
* update email field type to string in customer models and adjust related tests ([888646f](https://github.com/beaesthetic/beaesthetic-backend/commit/888646f1e9472fefd0cdaab363c9d2c91e579fd0))


### Bug Fixes

* update fidelity card API endpoints to remove trailing slashes ([1eee7b5](https://github.com/beaesthetic/beaesthetic-backend/commit/1eee7b55c2295dc5e7d5e506a723bb940bfe61c7))

## [1.12.0](https://github.com/beaesthetic/beaesthetic-backend/compare/customer-service-v1.11.0...customer-service-v1.12.0) (2026-07-09)


### Features

* add backfill command for MongoDB data to PostgreSQL and update config for MongoDB settings ([bfad929](https://github.com/beaesthetic/beaesthetic-backend/commit/bfad92972d5c639374be24828a9ec0c00cc71aa6))
* integrate Redis caching for customer data and update related configurations ([ceec999](https://github.com/beaesthetic/beaesthetic-backend/commit/ceec9999dbb77709f430b646d1c2e15affb926b4))

## [1.11.0](https://github.com/beaesthetic/beaesthetic-backend/compare/customer-service-v1.10.1...customer-service-v1.11.0) (2026-07-07)


### Features

* add Helm values files for appointment, customer, notification, and scheduler services ([611f314](https://github.com/beaesthetic/beaesthetic-backend/commit/611f31424ffe45ac75de9028cf1da9450fb54958))
* update MongoDB database name to remove environment suffix in values-dev.yaml ([1c3106f](https://github.com/beaesthetic/beaesthetic-backend/commit/1c3106f4303cfe7ec89e52c353d809c6dd4ed896))

## [1.10.1](https://github.com/beaesthetic/beaesthetic-backend/compare/customer-service-v1.10.0...customer-service-v1.10.1) (2025-12-15)


### Bug Fixes

* gradle improve customers ([77da985](https://github.com/beaesthetic/beaesthetic-backend/commit/77da9850a6e3556d33374f89c9de272fad3007b1))
* gradle improve customers ([8ffe7f2](https://github.com/beaesthetic/beaesthetic-backend/commit/8ffe7f2b4ec02fc2eeb84585f93083750291b098))
