// +build k8srequired

package chartconfig

import (
	"bytes"
	"context"
	"html/template"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/setup"
)

func DeleteResources(ctx context.Context, config setup.Config) error {
	items := []string{"cnr-server", "apiextensions-chart-config-e2e-chart"}

	for _, item := range items {
		err := config.HelmClient.DeleteRelease(ctx, item, helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func ExecuteValuesTemplate(ccv e2etemplates.ApiextensionsChartConfigValues) (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return buf.String(), nil
}

func InstallResources(ctx context.Context, config setup.Config) error {
	err := initializeCNR(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func initializeCNR(ctx context.Context, config setup.Config) error {
	err := installCNR(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCNR(ctx context.Context, config setup.Config) error {
	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: config.Logger,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	tarball, err := a.PullChartTarball(ctx, "cnr-server-chart", "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	err = config.HelmClient.InstallReleaseFromTarball(ctx, tarball, "giantswarm", helm.ReleaseName("cnr-server"), helm.ValueOverrides([]byte("{}")), helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
