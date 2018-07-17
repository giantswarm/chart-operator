// +build k8srequired

package templates

// ChartConfigResourceValues values required by the apiextensions-chart-config-e2e-chart.
const ChartConfigResourceValues = `chart:
  name: "tb-chart"
  channel: "1-0-beta"
  namespace: "giantswarm"
  release: "tb-release"
`
