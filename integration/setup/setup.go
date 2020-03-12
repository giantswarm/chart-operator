// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/spf13/afero"

	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/key"
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
		err := config.K8s.EnsureNamespaceCreated(ctx, key.Namespace())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// TODO: Use project.Version() once the operator is flattened.
	//
	//	https://github.com/giantswarm/giantswarm/issues/7896
	//
	var chartOperatorLatestRelease string
	{
		chartOperatorLatestRelease, err = appcatalog.GetLatestVersion(ctx, key.DefaultCatalogStorageURL(), project.Name())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var operatorTarballPath string
	{
		operatorVersion := fmt.Sprintf("%s-%s", chartOperatorLatestRelease, env.CircleSHA())
		operatorTarballURL, err := appcatalog.NewTarballURL(key.DefaultTestCatalogStorageURL(), project.Name(), operatorVersion)
		if err != nil {
			return microerror.Mask(err)
		}

		operatorTarballPath, err = config.HelmClient.PullChartTarball(ctx, operatorTarballURL)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		defer func() {
			fs := afero.NewOsFs()
			err := fs.Remove(operatorTarballPath)
			if err != nil {
				config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", operatorTarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		opts := helmclient.InstallOptions{
			ReleaseName: project.Name(),
		}
		err = config.HelmClient.InstallReleaseFromTarball(ctx, operatorTarballPath, key.Namespace(), map[string]interface{}{}, opts)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
