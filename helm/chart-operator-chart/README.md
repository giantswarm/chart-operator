# chart-operator Helm Chart
Helm Chart for the chart-operator

## Configuration

| Parameter          | Description                                 | Default                                                           |
|--------------------|---------------------------------------------|-------------------------------------------------------------------|
| `name`             | The name of the operator                    | `chart-operator`                                                  |
| `namespace`        | The namespaces of the operator              | `kube-system`                                                     |
| `port`             | The port of the operator container          | `8000`                                                            |
| `image.repository` | The operator container image repository     | `quay.io/giantswarm/chart-operator`                               |
| `image.tag`        | The operator container image tag            | `[latest commit SHA]`                                             |
| `resources`        | The operator pod resource requests & limits | `request: cpu:100m, memory:75Mi;  limits: cpu:250m, memory:250Mi` |

