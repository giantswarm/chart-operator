//go:build k8srequired
// +build k8srequired

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarE2EKubeconfig is the process environment variable representing the
	// E2E_KUBECONFIG env var.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"
)

var (
	buildVersion  string
	circleCI      string
	keepResources string
	kubeconfig    string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	filePath := filepath.Join(os.Getenv("CIRCLE_WORKING_DIRECTORY"), ".build_version")
	buf, err := os.ReadFile(filePath) // #nosec G304 G703 -- the CI working directory, not user input
	if err != nil {
		panic(fmt.Sprintf("error reading .build_version: %v", err))
	}

	buildVersion = strings.TrimSpace(string(buf))
	if buildVersion == "" {
		panic(".build_version must not be empty")
	}

	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfig == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarE2EKubeconfig))
	}
}

// BuildVersion is the version push-to-registries computed for this build. The
// chart is pushed to the test catalog with it as both version and appVersion.
func BuildVersion() string {
	return buildVersion
}

func CircleCI() string {
	return circleCI
}

func KeepResources() string {
	return keepResources
}

func KubeConfigPath() string {
	return kubeconfig
}
