package chartconfig

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"

	"github.com/giantswarm/chart-operator/integration/env"
)

func VersionBundleVersion(githubToken, testedVersion string) string {
	if githubToken == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", env.EnvVarGithubBotToken))
	}
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", env.EnvVarTestedVersion))
	}

	params := &framework.VBVParams{
		Component: "chart-operator",
		Provider:  "aws",
		Token:     githubToken,
		VType:     testedVersion,
	}
	versionBundleVersion, err := framework.GetVersionBundleVersion(params)
	if err != nil {
		panic(err.Error())
	}

	if versionBundleVersion == "" {
		if strings.ToLower(testedVersion) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}
		panic("version bundle version  must not be empty")
	}

	return versionBundleVersion
}
