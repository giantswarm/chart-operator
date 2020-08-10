# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## Changed

- Updated Kubernetes dependencies to v1.18.5.
- Updated Helm to v3.2.4.

## [1.0.7] - 2020-08-05

### Changed

- Rollback the helm release in pending-install, pending-upgrade.

## [1.0.6] - 2020-07-24

### Changed

- Disable force upgrades since recreating resources is not supported.
- Graduate Chart CRD to v1.
- Upgrade to operatorkit 1.2.0.

## [1.0.5] - 2020-07-15

### Changed

- Enable force upgrades when chart CR annotation is present.

## [1.0.4] - 2020-07-08

### Changed

- Update MD5 Hash only if chart-operator upgrade the release successfully. 
- Make kubernetes wait timeout configurable when installing and updating
releases.
- Set release revision in CR status.

## [v1.0.3] 2020-06-16

### Changed

- Fixed PodSecurityPolicy compatibility problem.

## [v1.0.2] 2020-06-04

### Changed

- Disabled force-upgrade from helmclient. 
- Canceling the release resource when migration is done yet.

## [v1.0.1] 2020-05-26

### Changed 

- Using helmclient v1.0.1 for security fix.
- Cancel the release resource when the manifest validation failed.

## [v1.0.0] 2020-05-18

### Changed

- Updated to support Helm 3; To keep using Helm 2, please use version 0.X.X.

## [v0.13.0] 2020-04-21

### Changed

- Deploy as a unique app in app collection in control plane clusters.

## [v0.12.4] 2020-04-15

### Changed

- Always set chart CR annotations so update state calculation is accurate.
- Only update failed Helm releases if the chart values or version has changed.

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

[Unreleased]: https://github.com/giantswarm/chart-operator/compare/v1.0.7...HEAD
[1.0.7]: https://github.com/giantswarm/chart-operator/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/giantswarm/chart-operator/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/giantswarm/chart-operator/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/giantswarm/chart-operator/compare/v1.0.3...v1.0.4
[v1.0.3]: https://github.com/giantswarm/chart-operator/compare/v1.0.2...v1.0.3
[v1.0.2]: https://github.com/giantswarm/chart-operator/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/giantswarm/chart-operator/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/giantswarm/chart-operator/compare/v0.13.0...v1.0.0
[v0.13.0]: https://github.com/giantswarm/chart-operator/compare/v0.12.4...v0.13.0
[v0.12.4]: https://github.com/giantswarm/chart-operator/compare/v0.12.3...v0.12.4
[v0.12.3]: https://github.com/giantswarm/chart-operator/compare/v0.12.2...v0.12.3
[v0.12.2]: https://github.com/giantswarm/chart-operator/compare/v0.12.1...v0.12.2
[v0.12.1]: https://github.com/giantswarm/chart-operator/compare/v0.12.0...v0.12.1
[v0.12.0]: https://github.com/giantswarm/chart-operator/compare/v0.8.0...v0.12.0
[v0.8.0]: https://github.com/giantswarm/chart-operator/compare/v0.7.0...v0.8.0

[v0.7.0]: https://github.com/giantswarm/chart-operator/releases/tag/v0.7.0
