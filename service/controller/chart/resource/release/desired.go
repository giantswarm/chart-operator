package release

import (
	"context"
	"crypto/md5" // #nosec
	"fmt"

	"github.com/imdario/mergo"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient/v3/pkg/helmclient"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
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
	err = mergo.Merge(&configMapData, secretData, mergo.WithOverride)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	convertFloat(configMapData)

	var valuesMD5Checksum string

	if len(configMapData) > 0 {
		// MD5 is only used for comparison but we need to turn off gosec or
		// linting errors will occur.
		h := md5.New() // #nosec
		_, err := h.Write([]byte(fmt.Sprintf("%v", configMapData)))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		valuesMD5Checksum = fmt.Sprintf("%x", h.Sum(nil))
	}

	releaseState := &ReleaseState{
		Name:              key.ReleaseName(cr),
		Status:            helmclient.StatusDeployed,
		ValuesMD5Checksum: valuesMD5Checksum,
		Values:            configMapData,
		Version:           key.Version(cr),
	}

	return releaseState, nil
}

func (r *Resource) getConfigMapData(ctx context.Context, cr v1alpha1.Chart) (map[string]interface{}, error) {
	configMapData := map[string]interface{}{}

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

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(ctx, configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "config map %#q in namespace %#q not found", configMapName, configMapNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, str := range configMap.Data {
			err := yaml.Unmarshal([]byte(str), &configMapData)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return configMapData, nil
}

func (r *Resource) getSecretData(ctx context.Context, cr v1alpha1.Chart) (map[string]interface{}, error) {
	var secretData map[string]interface{}

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

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret %#q in namespace %#q not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(secret.Data) != 1 {
			return nil, microerror.Mask(wrongTypeError)
		}

		for _, bytes := range secret.Data {
			err := yaml.Unmarshal(bytes, &secretData)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return secretData, nil
}
