package env

import (
	"os"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarKubeconfig is the process environment variable representing the
	// KUBECONFIG env var.
	EnvVarKubeconfig = "KUBECONFIG"
	// EnvVarE2EKubeconfig is the deprecated process environment variable
	// representing the E2E_KUBECONFIG env var. Replaced by KUBECONFIG.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"

	// e2eHarnessDefaultKubeconfig is defined to avoid dependency of
	// e2e-harness. e2e-harness depends on this project. We don't want
	// circular dependencies even though it works in this case. This makes
	// vendoring very tricky.
	//
	// NOTE this should reflect value of DefaultKubeConfig constant.
	//
	//	See https://godoc.org/github.com/giantswarm/e2e-harness/pkg/harness#pkg-constants.
	//
	// There is also a note in the code there.
	//
	//	See https://github.com/giantswarm/e2e-harness/pull/177
	//
	e2eHarnessDefaultKubeconfig = "/workdir/.shipyard/config"
)

var (
	circleCI      string
	circleSHA     string
	keepResources string
	kubeconfig    string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	circleSHA = os.Getenv(EnvVarCircleSHA)
	keepResources = os.Getenv(EnvVarKeepResources)

	kubeconfig = os.Getenv(EnvVarKubeconfig)
	if kubeconfig == "" {
		// EnvVarE2EKubeconfig is deprecated. We fall back to it if
		// EnvVarKubeconfig is not set.
		kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
		if kubeconfig == "" {
			// If neither env var is set we fall back to the e2e-harness
			// default location.
			kubeconfig = e2eHarnessDefaultKubeconfig
		}
	}
}

func CircleCI() string {
	return circleCI
}

func CircleSHA() string {
	return circleSHA
}

func KeepResources() string {
	return keepResources
}

func KubeConfigPath() string {
	return kubeconfig
}
