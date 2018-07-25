// +build k8srequired

package templates

// ChartOperatorResourceValues values required by chart-operator-resource-chart.
const ChartOperatorResourceValues = `chart:
  name: "tb-chart"
  channel: "5-5-beta"
  namespace: "giantswarm"
  release: "tb-release"
versionBundleVersion: 0.3.0
`

// UpdatedChartOperatorResourceValues values required to update chart-operator-resource-chart.
const UpdatedChartOperatorResourceValues = `chart:
  name: "tb-chart"
  channel: "5-6-beta"
  namespace: "giantswarm"
  release: "tb-release"
versionBundleVersion: 0.3.0
`
