package project

var (
	description = "chart-operator is our agent for deploying Helm charts as releases."
	name        = "chart-operator"
	gitSHA      = "n/a"
	source      = "https://github.com/giantswarm/chart-operator"
	version     = "n/a"
)

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
