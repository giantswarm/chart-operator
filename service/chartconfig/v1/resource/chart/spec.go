package chart

// ChartState holds the state of the chart to be reconciled.
type ChartState struct {
	// ChartName is the fully qualified name of the Helm Chart.
	// e.g. quay.io/giantswarm/chart-operator-chart
	ChartName string
	// ChannelName is the CNR channel to reconcile against.
	// e.g. 0.1-beta
	ChannelName string
	// ReleaseName is the Helm Release to be deployed.
	// e.g. 0.1.2
	ReleaseName string
}
