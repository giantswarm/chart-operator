// +build k8srequired

package templates

// ChartOperatorValues values required by chart-operator-chart.
const ChartOperatorValues = `cnr:
  address: http://cnr-server:5000
clusterDNSIP: 10.96.0.10
e2e: true

Installation:
  V1:
    Helm:
      HTTP:
        ClientTimeout: 30s
    Registry:
      Domain: quay.io
`
