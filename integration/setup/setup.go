// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/key"
	"github.com/giantswarm/chart-operator/integration/templates"
	"github.com/giantswarm/chart-operator/pkg/project"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	var v int
	var err error

	err = installResources(ctx, config)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "failed to install resources", "stack", fmt.Sprintf("%#v", err))
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	os.Exit(v)
}

func installResources(ctx context.Context, config Config) error {
	var err error

	{
		err = config.K8s.EnsureNamespaceCreated(ctx, key.Namespace())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = config.HelmClient.EnsureTillerInstalled(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var operatorTarballPath string
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "getting tarball URL")

		operatorTarballURL, err := appcatalog.GetLatestChart(ctx, key.DefaultTestCatalogStorageURL(), project.Name(), project.Version())
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("tarball URL is %#q", operatorTarballURL))

		config.Logger.LogCtx(ctx, "level", "debug", "message", "pulling tarball")

		operatorTarballPath, err = config.HelmClient.PullChartTarball(ctx, operatorTarballURL)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("tarball path is %#q", operatorTarballPath))
	}

	{
		defer func() {
			fs := afero.NewOsFs()
			err := fs.Remove(operatorTarballPath)
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", operatorTarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installing %#q", project.Name()))

		err = config.HelmClient.InstallReleaseFromTarball(ctx,
			operatorTarballPath,
			key.Namespace(),
			helm.ReleaseName(key.ReleaseName()),
			helm.ValueOverrides([]byte(templates.ChartOperatorValues)),
			helm.InstallWait(true))
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("installed %#q", project.Name()))
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "ensuring chart CRD exists")

		// The operator will install the CRD on boot but we create chart CRs
		// in the tests so this ensures the CRD is present.
		err = config.K8sClients.CRDClient().EnsureCreated(ctx, v1alpha1.NewChartCRD(), backoff.NewMaxRetries(7, 1*time.Second))
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "ensured chart CRD exists")
	}

	return nil
}
