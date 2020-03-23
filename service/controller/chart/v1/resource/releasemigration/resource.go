package releasemigration

import (
	"context"
	"fmt"
	"github.com/giantswarm/backoff"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/giantswarm/appcatalog"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/integration/key"
)

const (
	Name = "releasemigrationv1"
)

type Config struct {
	// Dependencies.
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	TillerNamespace string
}

type Resource struct {
	// Dependencies.
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	// Settings.
	tillerNamespace string
}

func New(config Config) (*Resource, error) {
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	r := &Resource{
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) findHelmV2Releases(ctx context.Context) ([]string, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "OWNER", "TILLER"),
	}

	// Check whether helm 2 release configMaps still exist.
	cms, err := r.k8sClient.CoreV1().ConfigMaps(r.tillerNamespace).List(lo)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	hadReleases := map[string]bool{}
	for _, cm := range cms.Items {
		if _, ok := hadReleases[cm.Name]; !ok {
			hadReleases[cm.Name] = true
		}
	}

	releases := make([]string, 0, len(hadReleases))
	for k := range hadReleases {
		releases = append(releases, k)
	}

	return releases, nil
}

func (r *Resource) ensureReleasesMigrated(ctx context.Context) error {

	// 0. Found all dangling helm release v2
	releases, err := r.findHelmV2Releases(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	// 1. Install helm-2to3-migration app
	{
		var tarballPath string
		{
			tarballURL, err := appcatalog.GetLatestChart(ctx, key.DefaultCatalogStorageURL(), "helm-2to3-migration")
			if err != nil {
				return microerror.Mask(err)
			}

			tarballPath, err = r.helmClient.PullChartTarball(ctx, tarballURL)
			if err != nil {
				return microerror.Mask(err)
			}

			defer func() {
				fs := afero.NewOsFs()
				err := fs.Remove(tarballPath)
				if err != nil {
					r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
				}
			}()

			opts := helmclient.InstallOptions{
				ReleaseName: "helm-2to3-migration",
			}

			values := map[string]interface{}{
				"releases": releases,
				"tiller": map[string]string{
					"namespace": r.tillerNamespace,
				},
			}

			err = r.helmClient.InstallReleaseFromTarball(ctx, tarballPath, r.tillerNamespace, values, opts)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	// 2. Wait until all helm v2 release are deleted
	operation := func() error {
		releases, err := r.findHelmV2Releases(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(releases) > 0 {
			return microerror.Maskf(releasesNotDeletedError, "helm v2 releases not deleted: %#q", releases)
		}

		return nil
	}

	notify := func(err error, t time.Duration) {
		r.logger.Log("level", "debug", "message", "failed to deleted all helm v2 releases")
	}

	b := backoff.NewConstant(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	err = backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
