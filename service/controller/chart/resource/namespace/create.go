package namespace

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: key.NamespaceAnnotations(cr),
			Labels:      key.NamespaceLabels(cr),
			Name:        key.Namespace(cr),
		},
	}

	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	ns.Labels[label.ManagedBy] = project.Name()

	r.logger.Debugf(ctx, "creating namespace %#q", ns.Name)

	ch := make(chan error)

	go func() {
		_, err = r.k8sClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		close(ch)
	}()

	select {
	case <-ch:
		// Fall through.
	case <-time.After(r.k8sWaitTimeout):
		r.logger.Debugf(ctx, "timeout creating namespace %#q", key.Namespace(cr))
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	if apierrors.IsAlreadyExists(err) {
		r.logger.Debugf(ctx, "already created namespace %#q", key.Namespace(cr))

		err = r.ensureNamespaceUpdated(ctx, cr)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "created namespace %#q", key.Namespace(cr))

	return nil
}

func (r *Resource) ensureNamespaceUpdated(ctx context.Context, cr v1alpha1.Chart) error {
	namespace, err := r.k8sClient.CoreV1().Namespaces().Get(ctx, key.Namespace(cr), metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	updated := true

	if namespace.GetLabels() == nil {
		namespace.Labels = map[string]string{}
	}

	for k, v := range key.NamespaceLabels(cr) {
		if namespace.GetLabels()[k] != v {
			namespace.GetLabels()[k] = v
			updated = false
		}
	}

	if namespace.GetAnnotations() == nil {
		namespace.Annotations = map[string]string{}
	}

	for k, v := range key.NamespaceAnnotations(cr) {
		if namespace.GetAnnotations()[k] != v {
			namespace.GetAnnotations()[k] = v
			updated = false
		}
	}

	if updated {
		// no-op
		return nil
	}

	r.logger.Debugf(ctx, "updating namespace %#q", namespace.Name)

	_, err = r.k8sClient.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "updated namespace %#q", namespace.Name)

	return nil
}
