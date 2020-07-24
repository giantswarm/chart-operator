package project

var (
	description = "chart-operator is an agent for deploying Helm charts as releases."
	gitSHA      = "n/a"
	name        = "chart-operator"
	source      = "https://github.com/giantswarm/chart-operator"
	version     = "1.0.6"
)

// ChartVersion is fixed for chart CRs. This is because they exist in both
// control plane and tenant clusters and their version is not linked to a
// release. We may revisit this in future.
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
