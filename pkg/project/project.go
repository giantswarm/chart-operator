package project

var (
	description = "chart-operator is our agent for deploying Helm charts as releases."
	gitSHA      = "n/a"
	name        = "chart-operator"
	source      = "https://github.com/giantswarm/chart-operator"
	version     = "1.0.0-alpha.1"
)

func ChartVersion() string {
	return "1.0.0"
}

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
