[![CircleCI](https://circleci.com/gh/giantswarm/chart-operator.svg?style=shield)](https://circleci.com/gh/giantswarm/chart-operator)

# chart-operator
test
The chart-operator deploys Helm charts as [helm] releases. It is implemented
using [operatorkit].

## Branches

- `master`
    - Latest version using Helm 3.
- `helm2`
    - Legacy support for Helm 2.

## chart CR

The operator deploys charts hosted in a Helm repository. The chart CRs are
managed by [app-operator] which provides a higher level abstraction for
managing apps via the app CRD.

### Example chart CR

```yaml
apiVersion: application.giantswarm.io/v1alpha1
kind: Chart
metadata:
  name: "prometheus"
  labels:
    chart-operator.giantswarm.io/version: "1.0.0"
spec:
  name: "prometheus"
  namespace: "monitoring"
  config:
    configMap:
      name: "prometheus-values"
      namespace: "monitoring"
    secret:
      name: "prometheus-secrets"
      namespace: "monitoring"
  tarballURL: "https://giantswarm.github.io/app-catalog/prometheus-1-0-0.tgz"
```

## Getting Project

Clone the git repository: https://github.com/giantswarm/chart-operator.git

### How to build

Build it using the standard `go build` command.

```
go build github.com/giantswarm/chart-operator
```

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/chart-operator/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

chart-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for
details.



[app-operator]: https://github.com/giantswarm/app-operator
[helm]: https://github.com/helm/helm
[operatorkit]: https://github.com/giantswarm/operatorkit
