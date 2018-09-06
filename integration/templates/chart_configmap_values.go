// +build k8srequired

package templates

// ChartConfigMapValues values required by the chart values configmap
const ChartConfigMapValues = `valuesJson: '{"config":"test-value"}'

secretJson: '{"secret":"test-secret"}'
`
