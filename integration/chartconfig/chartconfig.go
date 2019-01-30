// +build k8srequired

package chartconfig

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/integration/env"
)

func ExecuteChartConfigValuesTemplate(ccv e2etemplates.ApiextensionsChartConfigValues) (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return buf.String(), nil
}

func VersionBundleVersion(githubToken, testedVersion string) (string, error) {
	if githubToken == "" {
		return "", microerror.Maskf(invalidConfigError, "env var %#q must not be empty", env.EnvVarGithubBotToken)
	}
	if testedVersion == "" {
		return "", microerror.Maskf(invalidConfigError, "env var %#q must not be empty", env.EnvVarTestedVersion)
	}

	params := &framework.VBVParams{
		Component: "chart-operator",
		Provider:  "aws",
		Token:     githubToken,
		VType:     testedVersion,
	}
	versionBundleVersion, err := framework.GetVersionBundleVersion(params)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if versionBundleVersion == "" {
		if strings.ToLower(testedVersion) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}

		return "", microerror.Maskf(failedExecutionError, "version bundle version must not be empty")
	}

	return versionBundleVersion, nil
}
