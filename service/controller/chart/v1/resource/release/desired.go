package release

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	yaml "gopkg.in/yaml.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)
	tarballURL := key.TarballURL(cr)

	tarballPath, err := r.helmClient.PullChartTarball(ctx, tarballURL)
	if helmclient.IsPullChartFailedError(err) {
		// Add the status to the controller context. It will be used to set the
		// CR status in the status resource.
		cc.Status = controllercontext.Status{
			Reason: fmt.Sprintf("Pulling chart %#q failed", tarballURL),
			Release: controllercontext.Release{
				Status: releaseNotInstalledStatus,
			},
		}

		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("pulling chart %#q failed", tarballURL), "stack", microerror.Stack(err))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if helmclient.IsPullChartNotFound(err) {
		// Add the status to the controller context. It will be used to set the
		// CR status in the status resource.
		cc.Status = controllercontext.Status{
			Reason: fmt.Sprintf("Chart %#q not found", tarballURL),
			Release: controllercontext.Release{
				Status: releaseNotInstalledStatus,
			},
		}

		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("chart %#q not found", tarballURL), "stack", microerror.Stack(err))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if helmclient.IsPullChartTimeout(err) {
		// Add the status to the controller context. It will be used to set the
		// CR status in the status resource.
		cc.Status = controllercontext.Status{
			Reason: fmt.Sprintf("Chart %#q timeout", tarballURL),
			Release: controllercontext.Release{
				Status: releaseNotInstalledStatus,
			},
		}

		r.logger.LogCtx(ctx, "level", "warning", "message", fmt.Sprintf("chart %#q timeout", tarballURL), "stack", microerror.Stack(err))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	defer func() {
		err := r.fs.Remove(tarballPath)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
		}
	}()

	chart, err := r.helmClient.LoadChart(ctx, tarballPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapData, err := r.getConfigMapData(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretData, err := r.getSecretData(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Merge configmap and secret to provide a single set of values to Helm.
	values, err := helmclient.MergeValues(configMapData, secretData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var valuesYAML []byte
	var valuesMD5Checksum string

	if len(values) > 0 {
		valuesYAML, err = yaml.Marshal(values)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		valuesMD5Checksum = fmt.Sprintf("%x", md5.Sum(valuesYAML))
	} else {
		// We need to pass empty values in ValueOverrides to make the install
		// process use the default values and prevent errors on nested values.
		valuesYAML = []byte("{}")
	}

	releaseState := &ReleaseState{
		Name:              releaseName,
		Status:            helmDeployedStatus,
		ValuesMD5Checksum: valuesMD5Checksum,
		ValuesYAML:        valuesYAML,
		Version:           chart.Version,
	}

	return releaseState, nil
}

func (r *Resource) getConfigMapData(ctx context.Context, cr v1alpha1.Chart) (map[string][]byte, error) {
	configMapData := map[string][]byte{}

	// TODO: Improve desired state generation by removing call to key.IsDeleted.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/5719
	//
	if key.IsDeleted(cr) {
		// Return early as configmap has already been deleted.
		return configMapData, nil
	}

	if key.ConfigMapName(cr) != "" {
		configMapName := key.ConfigMapName(cr)
		configMapNamespace := key.ConfigMapNamespace(cr)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "config map %#q in namespace %#q not found", configMapName, configMapNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		// Convert strings to byte arrays to match secret types.
		for k, v := range configMap.Data {
			configMapData[k] = []byte(v)
		}
	}

	return configMapData, nil
}

func (r *Resource) getSecretData(ctx context.Context, cr v1alpha1.Chart) (map[string][]byte, error) {
	secretData := map[string][]byte{}

	// TODO: Improve desired state generation by removing call to key.IsDeleted.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/5719
	//
	if key.IsDeleted(cr) {
		// Return early as secret has already been deleted.
		return secretData, nil
	}

	if key.SecretName(cr) != "" {
		secretName := key.SecretName(cr)
		secretNamespace := key.SecretNamespace(cr)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret %#q in namespace %#q not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		secretData = secret.Data
	}

	return secretData, nil
}
