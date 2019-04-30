package env

import (
	"fmt"
	"os"
)

const (
	EnvVarCircleCI       = "CIRCLECI"
	EnvVarCircleSHA      = "CIRCLE_SHA1"
	EnvVarE2EKubeconfig  = "E2E_KUBECONFIG"
	EnvVarGithubBotToken = "GITHUB_BOT_TOKEN"
	EnvVarKeepResources  = "KEEP_RESOURCES"
	EnvVarTestedVersion  = "TESTED_VERSION"
)

var (
	circleCI      string
	circleSHA     string
	githubToken   string
	keepResources string
	kubeconfig    string
	testedVersion string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)
	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	// Optional environment variables only needed for chartconfig tests.
	githubToken = os.Getenv(EnvVarGithubBotToken)
	testedVersion = os.Getenv(EnvVarTestedVersion)
}

func CircleCI() string {
	return circleCI
}

func CircleSHA() string {
	return circleSHA
}

func GithubToken() string {
	return githubToken
}

func KeepResources() string {
	return keepResources
}

func KubeConfigPath() string {
	return kubeconfig
}

func TestedVersion() string {
	return testedVersion
}
