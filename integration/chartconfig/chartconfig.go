package chartconfig

import (
	"bytes"
	"fmt"
	"html/template"

	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

type ChartConfigValues struct {
	Channel              string
	ConfigMap            ConfigMap
	Name                 string
	Namespace            string
	Release              string
	Secret               Secret
	VersionBundleVersion string
}

type ConfigMap struct {
	Name            string
	Namespace       string
	ResourceVersion string
}

type Secret struct {
	Name            string
	Namespace       string
	ResourceVersion string
}

func (ccv ChartConfigValues) ExecuteChartValuesTemplate() (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (ccv ChartConfigValues) UpdateChartOperatorResource(logger micrologger.Logger, helmClient *helmclient.Client, releaseName string) error {
	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: logger,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	tarballPath, err := a.PullChartTarball(fmt.Sprintf("%s-chart", releaseName), "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	chartValues, err := ccv.ExecuteChartValuesTemplate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = helmClient.UpdateReleaseFromTarball(releaseName, tarballPath,
		helm.UpdateValueOverrides([]byte(chartValues)))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
