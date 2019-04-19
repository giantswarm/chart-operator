package release

// ReleaseState holds the state of the Helm release to be reconciled.
type ReleaseState struct {
	// Name is the name of the Helm release when the chart is deployed.
	// e.g. chart-operator
	Name string
	// Status is the status of the Helm release when the chart is deployed.
	// e.g. DEPLOYED
	Status string
	// ValuesMD5Checksum is the MD5 checksum of the values YAML. It is used for
	// comparison since it is more reliable than using the values returned by
	// helmclient.GetReleaseContent.
	ValuesMD5Checksum string
	// ValuesYAML are any values that have been set when the Helm Chart was
	// installed.
	ValuesYAML []byte
	// Version is the version of the Helm Chart to be deployed.
	// e.g. 0.1.2
	Version string
}
