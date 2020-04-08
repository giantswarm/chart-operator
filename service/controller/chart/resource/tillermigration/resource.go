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
	Name = "tillermigration"
)

type Config struct {
	// Dependencies.
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	TillerNamespace string
}

type Resource struct {
	// Dependencies.
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// Settings.
	tillerNamespace string
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

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	r := &Resource{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ensureTillerDeleted(ctx context.Context) error {
	name := "tiller-giantswarm"

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting service account %#q in namespace %#q", name, r.tillerNamespace))
	err := r.k8sClient.CoreV1().ServiceAccounts(r.tillerNamespace).Delete(name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted service account %#q in namespace %#q", name, r.tillerNamespace))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting cluster role binding %#q", name))
	err = r.k8sClient.RbacV1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted cluster role binding %#q", name))

	pspName := fmt.Sprintf("%s-psp", name)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting psp cluster role binding %#q", pspName))
	err = r.k8sClient.RbacV1().ClusterRoleBindings().Delete("tiller-giantswarm-psp", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted psp cluster role binding %#q", pspName))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting psp cluster role %#q", pspName))
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted psp cluster role %#q", pspName))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting pod security policy %#q", pspName))
	err = r.k8sClient.PolicyV1beta1().PodSecurityPolicies().Delete(pspName, &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted pod security policy %#q", pspName))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting network policy %#q in namespace %#q", pspName, r.tillerNamespace))
	err = r.k8sClient.NetworkingV1().NetworkPolicies(r.tillerNamespace).Delete("tiller-giantswarm", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted network policy %#q in namespace %#q", pspName, r.tillerNamespace))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting deployment `tiller-deploy` in namespace %#q", r.tillerNamespace))
	err = r.k8sClient.AppsV1().Deployments(r.tillerNamespace).Delete("tiller-deploy", &metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// no-op
	} else if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted deployment `tiller-deploy` in namespace %#q", r.tillerNamespace))

	return nil
}
