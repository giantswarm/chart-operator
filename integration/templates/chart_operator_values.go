//go:build k8srequired
// +build k8srequired

package templates

// ChartOperatorValues values required by chart-operator-chart.
const ChartOperatorValues = `cnr:
  address: http://cnr-server:5000
clusterDNSIP: 10.96.0.10
e2e: false

isManagementCluster: true
registry:
  domain: quay.io

verticalPodAutoscaler:
  enabled: false
`
