package chart

// ChartState holds the state of the chart to be reconciled.
type ChartState struct {
	// ChartValues are any values that have been set when the Helm Chart was
	// installed.
	ChartValues map[string]interface{}
	// ReleaseName is the name of the Helm release when the chart is deployed.
	// e.g. chart-operator
	ReleaseName string
	// ReleaseStatus is the status of the Helm release when the chart is deployed.
	// e.g. DEPLOYED
	ReleaseStatus string
	// ReleaseVersion is the version of the Helm Chart to be deployed.
	// e.g. 0.1.2
	ReleaseVersion string
}
