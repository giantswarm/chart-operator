package controller

const (
	// ConfigMapValuesKey is the key name for values configmaps that are
	// applied when installing or updating charts.
	ConfigMapValuesKey = "values.json"
	// SecretValuesKey is the key name for values secrets that should be applied
	// when installing or updating charts.
	SecretValuesKey = "secret.json"
)
