package release

import (
	"context"
	"crypto/md5" // #nosec
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/chart-operator/service/controller/chart/key"
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
	values, err := helmclient.MergeValues(configMapData, secretData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var valuesYAML []byte
	var valuesMD5Checksum string

	if len(values) > 0 {
		// We serialize the values to YAML so we can generate the MD5 checksum.
		// We use this for comparison because Helm may modify the values we get
		// back from the Helm client.
		valuesYAML, err = yaml.Marshal(values)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// MD5 is only used for comparison but we need to turn off gosec or
		// linting errors will occur.
		valuesMD5Checksum = fmt.Sprintf("%x", md5.Sum(valuesYAML)) // #nosec
	} else {
		// We need to pass empty values in ValueOverrides to make the install
		// process use the default values and prevent errors on nested values.
		valuesYAML = []byte("{}")
	}

	releaseState := &ReleaseState{
		Name:              key.ReleaseName(cr),
		Status:            helmclient.StatusDeployed,
		ValuesMD5Checksum: valuesMD5Checksum,
		Values:            values,
		Version:           key.Version(cr),
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
