package chart

// ChartState holds the state of the chart to be reconciled.
type ChartState struct {
	// ChartName is the fully qualified name of the Helm Chart.
	// e.g. quay.io/giantswarm/chart-operator-chart
	ChartName string
	// ChartValues are any values that have been set when the Helm Chart was
	// installed.
	ChartValues map[string]interface{}
	// ChannelName is the CNR channel to reconcile against.
	// e.g. 0.1-beta
	ChannelName string
	// ReleaseName is the name of the Helm release when the chart is deployed.
	// e.g. chart-operator
	ReleaseName string
	// ReleaseStatus is the status of the Helm Release.
	// e.g. DEPLOYED
	ReleaseStatus string
	// ReleaseVersion is the version of the Helm Chart to be deployed.
	// e.g. 0.1.2
	ReleaseVersion string
}

// Equals asseses the equality of ChartStates with regards to distinguishing fields.
func (a *ChartState) Equals(b ChartState) bool {
	if a.ReleaseName != b.ReleaseName {
		return false
	}
	if a.ReleaseVersion != b.ReleaseVersion {
		return false
	}
	if a.ChartName != b.ChartName {
		return false
	}
	return true
}
