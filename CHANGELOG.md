# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

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

## [0.8.0]

### Added

- Separate network policy.

## [0.7.0]

### Added

- Separate podsecuritypolicy.
- Security context in deployment spec with non-root user.
