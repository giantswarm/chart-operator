# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.12.3] 2020-04-09

### Changed

- Fix problem pushing chart to default app catalog.

### Changed

## [v0.12.2] 2020-04-09

### Changed

- Fix update state calculation and status resource for long running deployments.
- Handle 503 responses when GitHub Pages is unavailable.
- Make HTTP client timeout configurable for pulling chart tarballs in AWS China.
- Switch from dep to go modules.

## [v0.12.1] 2020-03-10

### Removed

- Remove usage of legacy chartconfig CRs in Tiller metrics.

## [v0.12.0] 2020-03-09

### Added

- Add chartmigration resource to allow legacy chartconfig controller to be
removed. ([#358](https://github.com/giantswarm/chart-operator/pull/358))

### Changed

- Improve reason field in chart CR status when installing a chart fails. ([#359](https://github.com/giantswarm/chart-operator/pull/359))
- Use version from chart CR when calculating desired state to reduce number of
HTTP requests to pull chart tarballs. ([#351](https://github.com/giantswarm/chart-operator/pull/353))
- Wait for deleted Helm release before removing finalizer. ([#360](https://github.com/giantswarm/chart-operator/pull/360))
- Do not wait when installing or updating a Helm release takes over 3 seconds.
We check progress in the next reconciliation loop. ([#362](https://github.com/giantswarm/chart-operator/pull/362))

### Removed

- Remove legacy chartconfig controller. ([#365](https://github.com/giantswarm/chart-operator/pull/365))

## [v0.8.0]

### Added

- Separate network policy.

## [v0.7.0]

### Added

- Separate podsecuritypolicy.
- Security context in deployment spec with non-root user.

[Unreleased]: https://github.com/giantswarm/chart-operator/compare/v0.12.1..HEAD
[v0.12.1]: https://github.com/giantswarm/chart-operator/releases/tag/v0.12.1
[v0.12.0]: https://github.com/giantswarm/chart-operator/releases/tag/v0.12.0
[v0.8.0]: https://github.com/giantswarm/chart-operator/releases/tag/v0.8.0
[v0.7.0]: https://github.com/giantswarm/chart-operator/releases/tag/v0.7.0
