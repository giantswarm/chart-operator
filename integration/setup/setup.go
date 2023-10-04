//go:build k8srequired
// +build k8srequired

package setup

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"

	"github.com/giantswarm/chart-operator/v3/integration/env"
	"github.com/giantswarm/chart-operator/v3/integration/key"
	"github.com/giantswarm/chart-operator/v3/pkg/project"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	var v int
	var err error

	err = installResources(ctx, config)
	if err != nil {
		config.Logger.Errorf(ctx, err, "failed to install resources")
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

	var operatorTarballURL string
	{
		config.Logger.Debugf(ctx, "getting %#q tarball URL", project.Name())

		o := func() error {
			operatorTarballURL, err = appcatalog.GetLatestChart(ctx, key.DefaultTestCatalogStorageURL(), project.Name(), env.CircleSHA())
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		b := backoff.NewConstant(5*time.Minute, 10*time.Second)
		n := backoff.NewNotifier(config.Logger, ctx)

		err = backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.Debugf(ctx, "tarball URL is %#q", operatorTarballURL)
	}

	var operatorTarballPath string
	{
		config.Logger.Debugf(ctx, "pulling tarball")

		operatorTarballPath, err = config.HelmClient.PullChartTarball(ctx, operatorTarballURL)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.Debugf(ctx, "tarball path is %#q", operatorTarballPath)
	}

	{
		defer func() {
			fs := afero.NewOsFs()
			err := fs.Remove(operatorTarballPath)
			if err != nil {
				config.Logger.Errorf(ctx, err, "deletion of %#q failed", operatorTarballPath)
			}
		}()

		config.Logger.Debugf(ctx, "installing %#q", project.Name())

		opts := helmclient.InstallOptions{
			ReleaseName: project.Name(),
		}
		values := map[string]interface{}{
			"clusterDNSIP": "10.96.0.10",
		}
		err = config.HelmClient.InstallReleaseFromTarball(ctx, operatorTarballPath, key.Namespace(), values, opts)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.Debugf(ctx, "installed %#q", project.Name())
	}

	return nil
}
