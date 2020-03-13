package tillermigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "tillermigrationv1"
)

type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type Resource struct {
	// Dependencies.
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureTillerDeleted(ctx context.Context) error {
	name := "tiller-giantswarm"
	namespace := "giantswarm"

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting serviceAccount %#q in namespace %#q", name, namespace))
	err := r.k8sClient.CoreV1().ServiceAccounts(namespace).Delete(name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted serviceAccount %#q in namespace %#q", name, namespace))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting clusterRoleBinding %#q", name))
	err = r.k8sClient.RbacV1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted clusterRoleBinding %#q", name))

	pspName := fmt.Sprintf("%s-psp", name)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting clusterRoleBinding %#q", pspName))
	err = r.k8sClient.RbacV1().ClusterRoleBindings().Delete("tiller-giantswarm-psp", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted clusterRoleBinding %#q", pspName))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting clusterRole %#q", pspName))
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted clusterRole %#q", pspName))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting podSecurityPolicy %#q", pspName))
	err = r.k8sClient.PolicyV1beta1().PodSecurityPolicies().Delete(pspName, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted podSecurityPolicy %#q", pspName))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting networkPolicy %#q in namespace %#q", pspName, namespace))
	err = r.k8sClient.NetworkingV1().NetworkPolicies(namespace).Delete("tiller-giantswarm", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted networkPolicy %#q in namespace %#q", pspName, namespace))

	r.logger.LogCtx(ctx, "level", "debug", "message", "deleting priorityClass `giantswarm-critical`")
	err = r.k8sClient.SchedulingV1().PriorityClasses().Delete("giantswarm-critical", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleted priorityClass `giantswarm-critical`")

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting deployment `tiller-deploy` in namespace %#q", namespace))
	err = r.k8sClient.AppsV1().Deployments(namespace).Delete("tiller-deploy", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted deployment `tiller-deploy` in namespace %#q", namespace))

	return nil
}
