# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project's packages adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Move pss values under the global property

### Changed

- Use base images from `gsoci.azurecr.io`

## [3.1.2] - 2023-12-20

### Changed

- Configure `gsoci.azurecr.io` as the default container image registry.

## [3.1.1] - 2023-12-05

### Changed

- Configure gsoci.azurecr.io as the registry to use by default

## [3.1.0] - 2023-10-19

### Changed

- Force-disable PSP-related resources when `global.podSecurityStandards.enforced` value is true.

## [3.0.0] - 2023-10-04

### Removed

- Removed `giantswarm.io/monitoring: "true"` label from the `Service` resource. To get metrics `chart-operator` should
  be from now on used in conjunction with `chart-operator-extensions` version `v1.1.1` or later to deploy
  `ServiceMonitor` resource for it. It was split up as `chart-operator` is one of the first component to get into
  a cluster that will deploy most other things, for example Prometheus that will eventually actually deploy the
  CRD for `ServiceMonitor`.

## [2.35.2] - 2023-09-26

### Fixed

- Fixed default values for `.proxy` and `.cluster.proxy` values and updated Helm values schema accordingly.

## [2.35.1] - 2023-09-21

### Changed

- Changed pod taint toleration to only tolerate `NotReady` for CAPI.

## [2.35.0] - 2023-05-04

### Changed

- Disable PSPs for k8s 1.25 and newer.

## [2.34.1] - 2023-05-02

### Added

- Add Kyverno `PolicyExceptions` for necessary capabilities normally prohibited by PSS policies.

## [2.34.0] - 2023-02-14

### Changed

- Selecting private Helm client on demand for some operations.

## [2.33.2] - 2022-12-16

## [2.33.1] - 2022-12-16

### Added

- New error for values schema validation.

### Changed

- Use transitional errors coming from running Helm in the Chart CR status.

## [2.33.0] - 2022-11-16

### Added

- Add support to run in private cloud clusters, which cannot provide any working `externalDNSIP`.

## [2.32.0] - 2022-11-15

## Added

- Support for running behind a proxy.
  - `HTTP_PROXY`,`HTTPS_PROXY` and `NO_PROXY` are set as environment variables in `deployment/chart-operator` if defined in `values.yaml`.
- Support for using `cluster-apps-operator` generated `cluster.proxy` values.

## [2.31.0] - 2022-10-07

## Added

- Add internal upgrade step on installation for Helm charts marked by annotation.

## [2.30.0] - 2022-09-23

### Added

- Add suport for timeouts fields in the Chart CR.

### Changed

- Add support for new control-plane label in k8s 1.24.

## [2.29.0] - 2022-08-12

- Reconfigure VPA autoscaler to react correctly to pod resource ceilings

## [2.28.0] - 2022-08-09

### Changed

- Add `pre-upgrade` helm annotation to `giantswarm-critical` PriorityClass in order to fix upgrade issues.

## [2.27.0] - 2022-07-29

### Added

- Ensure the `giantswarm-critical` PriorityClass is created first on initial installation.

## [2.26.0] - 2022-07-20

### Changed

- Use `127.0.0.1` as KUBERNETES_SERVICE_HOST when `bootstrapMode` is enabled.

## [2.25.0] - 2022-07-04

### Changed

- Tighten pod and container security contexts for PSS restricted policies.
- Use downward API to set deployment env var `KUBERNETES_SERVICE_HOST` to `status.hostIP`.
- Change `initialBootstrapMode` configuration value to `bootstrapMode`.
- Use private Helm client for installing app-operators from control-plane-test-catalog

### Added

- Allow to set api server pod port when enabling `initialBootstrapMode`.

## [2.24.1] - 2022-06-22

### Changed

- Update `helmclient` to v4.10.1.

## [2.24.0] - 2022-06-09

### Changed

- Add `chart-pull-failed` error to differentiate between issues when pulling chart tarball and other problems.

### Fixed

- Fix missing `PriorityClass` issue.

## [2.23.0] - 2022-06-06

### Changed

- Always create `giantswarm-critical` priority class if it does not exist.
- Add initialBootstrapMode flag to allow deploying CNI as managed apps.

## [2.22.0] - 2022-05-30

### Added

- Split Helm client into private Helm client for `giantswarm`-namespaced apps and public Helm client for rest of the apps.

## [2.21.1] - 2022-05-19

### Added

- Add Helm release failure reason when it is known, and if there is a currently successfully released version

## [2.21.0] - 2022-04-07

### Changed

- Update `helmclient` to v4.10.0.

## [2.20.1] - 2022-03-15

### Changed

- Use `apptestctl` to install CRDs in integration tests to avoid hitting GitHub rate limits.

### Fixed

- Fix `status` resource to use Helm release status if it exists.

## [2.20.0] - 2021-12-15

### Changed

- Update Helm to v3.6.3.
- Use controller-runtime client to remove CAPI dependency.

### Removed

- Remove unused helm 2 release collector.

## [2.19.1] - 2021-10-20

### Changed

- Deployment `hostNetwork` is enabled or not depending on `chartOperator.cni.install` value.

## [2.19.0] - 2021-08-13

### Removed

- Remove `tillermigration` resource now Helm 3 migration is complete.

## [2.18.1] - 2021-08-05

### Changed

- Increase memory limit for deploying large charts in workload clusters.

## [2.18.0] - 2021-06-21

## Added

- Add releasemaxhistory resource which ensures we retry at a reduced rate when
there are repeated failed upgrades.

### Changed

- Upgrade Helm release when failed even if version or values have not changed
to handle situations like failed webhooks where we should retry.

## [2.17.0] - 2021-06-09

### Changed

- Prepare helm values to configuration management.
- Update architect-orb to v3.0.0.

### Fixed

- Improve status message when helm release has failed max number of attempts.

## [2.16.0] - 2021-06-03

### Changed

For CAPI clusters:

- Add tolerations to start on `NotReady` nodes for installing CNI.
- Create `giantswarm-critical` priority class.
- Use host network to allow installing CNI packaged as an app.

## [2.15.0] - 2021-05-20

### Added

- Proxy support in helm template.

## [2.14.0] - 2021-04-30

### Changed

- Cancel the release resource when the manifest object already exists.
- Cancel the release resource when helm returns an unknown error.

## [2.13.1] - 2021-04-06

### Fixed

- Updated OperatorKit to v4.3.1 for Kubernetes 1.20 support.

## [2.13.0] - 2021-03-31

### Changed

- `giantswarm-critical` PriorityClass only managed when E2E.

## [2.12.0] - 2021-03-26

### Changed

- Set docker.io as the default registry
- Pass RESTMapper to helmclient to reduce the number of REST API calls.
- Updated Helm to v3.5.3.

## [2.11.0] - 2021-03-19

### Added

- Updating namespace metadata using namespaceConfig in `Chart` CRs.

## [2.10.0] - 2021-03-17

### Added

- Pause Chart CR reconciliation when it has chart-operator.giantswarm.io/paused=true annotation.

### Changed

- Deploy `giantswarm-critical` PriorityClass when it's not found.

## [2.9.0] - 2021-02-03

### Added

- Use diff key when logging differences between the current and desired release.

### Fixed

- Stop updating Helm release if it has failed the previous 5 attempts.

## [2.8.0] - 2021-01-27

### Added

- Add support for skip CRD flag when installing Helm releases.

## [2.7.1] - 2021-01-13

### Fixed

- Only create VPA if autoscaling API group is present.

## [2.7.0] - 2021-01-07

### Added

- Added last reconciled timestamp as metrics.

## [2.6.0] - 2020-12-21

### Added

- Print difference between current release and desired release.

### Changed

- Updated Helm to v3.4.2.

## [2.5.2] - 2020-12-07

### Added

- Add Vertical Pod Autoscaler support.

## [2.5.1] - 2020-12-01

### Fixed

- Fix comparison of last deployed and revision optional fields in status resource.
- Set memory limit and reduce requests.

## [2.5.0] - 2020-11-09

### Added

- Validate the cache in helmclient to avoid state requests when pulling tarballs.
- Call status webhook with token values.

### Fixed

- Update apiextensions to v3 and replace CAPI with Giant Swarm fork.

## [2.4.0] - 2020-10-29

### Added

- Call status webhook when webhook annotation is present.

### Removed

- Remove chartmigration resource as migration from chartconfig to chart CRs is
complete.

## [2.3.5] - 2020-10-13

### Fixed

- Stop repeating helm upgrade for the failed helm release.

## [2.3.4] - 2020-10-01

### Added

- Added release name as a label into the event count metrics.

## [2.3.3] - 2020-09-29

### Fixed

- Updated Helm to v3.3.4.
- Updated Kubernetes dependencies to v1.18.9.
- Update deployment annotation to use checksum instead of helm revision to
reduce how often pods are rolled.
- Increase wait timeout for accessing Kubernetes API from 10s to 120s.

## [2.3.2] - 2020-09-22

### Added

- Added event count metrics for delete, install, rollback and update of Helm releases.

### Fixed

- Fix structs merging error in helmclient.

### Security

- Updated Helm to v3.3.3.

## [2.3.1] - 2020-09-04

### Added

- Add monitoring labels.

### Changed

- Add namespace to logging message.

### Fixed

- Remove memory limits from deployment.

## [2.3.0] - 2020-08-24

### Changed

- Using default DNS policy for control planes.

## [2.2.1] - 2020-08-19

### Changed

- Fixed the timeout value for the namespace resource.

## [2.2.0] - 2020-08-19

### Changed

- Creating namespace before helm operations.

## [2.1.0] - 2020-08-18

## Changed

- Updated Helm to v3.3.0.

## [2.0.0] - 2020-08-12

## Changed

- Updated backward incompatible Kubernetes dependencies to v1.18.5.
- Updated Helm to v3.2.4.
- Fix the rollback in a loop problem.

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

[Unreleased]: https://github.com/giantswarm/chart-operator/compare/v3.1.2...HEAD
[3.1.2]: https://github.com/giantswarm/chart-operator/compare/v3.1.1...v3.1.2
[3.1.1]: https://github.com/giantswarm/chart-operator/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/giantswarm/chart-operator/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/giantswarm/chart-operator/compare/v2.35.2...v3.0.0
[2.35.2]: https://github.com/giantswarm/chart-operator/compare/v2.35.1...v2.35.2
[2.35.1]: https://github.com/giantswarm/chart-operator/compare/v2.35.0...v2.35.1
[2.35.0]: https://github.com/giantswarm/chart-operator/compare/v2.34.1...v2.35.0
[2.34.1]: https://github.com/giantswarm/chart-operator/compare/v2.34.0...v2.34.1
[2.34.0]: https://github.com/giantswarm/chart-operator/compare/v2.33.2...v2.34.0
[2.33.2]: https://github.com/giantswarm/chart-operator/compare/v2.33.1...v2.33.2
[2.33.1]: https://github.com/giantswarm/chart-operator/compare/v2.33.0...v2.33.1
[2.33.0]: https://github.com/giantswarm/chart-operator/compare/v2.32.0...v2.33.0
[2.32.0]: https://github.com/giantswarm/chart-operator/compare/v2.31.0...v2.32.0
[2.31.0]: https://github.com/giantswarm/chart-operator/compare/v2.30.0...v2.31.0
[2.30.0]: https://github.com/giantswarm/chart-operator/compare/v2.29.0...v2.30.0
[2.29.0]: https://github.com/giantswarm/chart-operator/compare/v2.28.0...v2.29.0
[2.28.0]: https://github.com/giantswarm/chart-operator/compare/v2.27.0...v2.28.0
[2.27.0]: https://github.com/giantswarm/chart-operator/compare/v2.26.0...v2.27.0
[2.26.0]: https://github.com/giantswarm/chart-operator/compare/v2.25.0...v2.26.0
[2.25.0]: https://github.com/giantswarm/chart-operator/compare/v2.24.1...v2.25.0
[2.24.1]: https://github.com/giantswarm/chart-operator/compare/v2.24.0...v2.24.1
[2.24.0]: https://github.com/giantswarm/chart-operator/compare/v2.23.0...v2.24.0
[2.23.0]: https://github.com/giantswarm/chart-operator/compare/v2.22.0...v2.23.0
[2.22.0]: https://github.com/giantswarm/chart-operator/compare/v2.21.1...v2.22.0
[2.21.1]: https://github.com/giantswarm/chart-operator/compare/v2.21.0...v2.21.1
[2.21.0]: https://github.com/giantswarm/chart-operator/compare/v2.20.1...v2.21.0
[2.20.1]: https://github.com/giantswarm/chart-operator/compare/v2.20.0...v2.20.1
[2.20.0]: https://github.com/giantswarm/chart-operator/compare/v2.19.1...v2.20.0
[2.19.1]: https://github.com/giantswarm/chart-operator/compare/v2.19.0...v2.19.1
[2.19.0]: https://github.com/giantswarm/chart-operator/compare/v2.18.1...v2.19.0
[2.18.1]: https://github.com/giantswarm/chart-operator/compare/v2.18.0...v2.18.1
[2.18.0]: https://github.com/giantswarm/chart-operator/compare/v2.17.0...v2.18.0
[2.17.0]: https://github.com/giantswarm/chart-operator/compare/v2.16.0...v2.17.0
[2.16.0]: https://github.com/giantswarm/chart-operator/compare/v2.15.0...v2.16.0
[2.15.0]: https://github.com/giantswarm/chart-operator/compare/v2.14.0...v2.15.0
[2.14.0]: https://github.com/giantswarm/chart-operator/compare/v2.13.1...v2.14.0
[2.13.1]: https://github.com/giantswarm/chart-operator/compare/v2.13.0...v2.13.1
[2.13.0]: https://github.com/giantswarm/chart-operator/compare/v2.12.0...v2.13.0
[2.12.0]: https://github.com/giantswarm/chart-operator/compare/v2.11.0...v2.12.0
[2.11.0]: https://github.com/giantswarm/chart-operator/compare/v2.10.0...v2.11.0
[2.10.0]: https://github.com/giantswarm/chart-operator/compare/v2.9.0...v2.10.0
[2.9.0]: https://github.com/giantswarm/chart-operator/compare/v2.8.0...v2.9.0
[2.8.0]: https://github.com/giantswarm/chart-operator/compare/v2.7.1...v2.8.0
[2.7.1]: https://github.com/giantswarm/chart-operator/compare/v2.7.0...v2.7.1
[2.7.0]: https://github.com/giantswarm/chart-operator/compare/v2.6.0...v2.7.0
[2.6.0]: https://github.com/giantswarm/chart-operator/compare/v2.5.2...v2.6.0
[2.5.2]: https://github.com/giantswarm/chart-operator/compare/v2.5.1...v2.5.2
[2.5.1]: https://github.com/giantswarm/chart-operator/compare/v2.5.0...v2.5.1
[2.5.0]: https://github.com/giantswarm/chart-operator/compare/v2.4.0...v2.5.0
[2.4.0]: https://github.com/giantswarm/chart-operator/compare/v2.3.5...v2.4.0
[2.3.5]: https://github.com/giantswarm/chart-operator/compare/v2.3.4...v2.3.5
[2.3.4]: https://github.com/giantswarm/chart-operator/compare/v2.3.3...v2.3.4
[2.3.3]: https://github.com/giantswarm/chart-operator/compare/v2.3.2...v2.3.3
[2.3.2]: https://github.com/giantswarm/chart-operator/compare/v2.3.1...v2.3.2
[2.3.1]: https://github.com/giantswarm/chart-operator/compare/v2.3.0...v2.3.1
[2.3.0]: https://github.com/giantswarm/chart-operator/compare/v2.2.1...v2.3.0
[2.2.1]: https://github.com/giantswarm/chart-operator/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/giantswarm/chart-operator/compare/v2.1.0...v2.2.0
[2.1.0]: https://github.com/giantswarm/chart-operator/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/giantswarm/chart-operator/compare/v1.0.7...v2.0.0
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
