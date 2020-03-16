// +build k8srequired

package templates

// ChartOperatorValues values required by chart-operator-chart.
const ChartOperatorValues = `cnr:
  address: http://cnr-server:5000
clusterDNSIP: 10.96.0.10
e2e: false
`
