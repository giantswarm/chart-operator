package env

import (
	"fmt"
	"os"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarGithubBotToken is the process environment variable representing
	// the GITHUB_BOT_TOKEN env var.
	EnvVarGithubBotToken = "GITHUB_BOT_TOKEN"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"
	// EnvVarTestedVersion is the process environment variable representing the
	// TESTED_VERSION env var.
	EnvVarTestedVersion = "TESTED_VERSION"
)

var (
	circleCI      string
	circleSHA     string
	githubToken   string
	keepResources string
	testedVersion string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

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

func TestedVersion() string {
	return testedVersion
}
