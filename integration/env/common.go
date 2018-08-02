package env

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"
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
	circleCI             string
	circleSHA            string
	keepResources        string
	testedVersion        string
	versionBundleVersion string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	var err error
	token := os.Getenv(EnvVarGithubBotToken)
	params := &framework.VBVParams{
		Component: "chart-operator",
		Provider:  "aws",
		Token:     token,
		VType:     TestedVersion(),
	}
	versionBundleVersion, err = framework.GetVersionBundleVersion(params)
	if err != nil {
		panic(err.Error())
	}

	if VersionBundleVersion() == "" {
		if strings.ToLower(TestedVersion()) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}
		panic("version bundle version  must not be empty")
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

func TestedVersion() string {
	return testedVersion
}

func VersionBundleVersion() string {
	return versionBundleVersion
}
