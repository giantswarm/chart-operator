package key

func TillerDeploymentName() string {
	return "tiller-deploy"
}

func TillerMaxHistoryEnvVarName() string {
	return "TILLER_HISTORY_MAX"
}

func TillerMaxHistoryEnvVarValue() string {
	return "10"
}
