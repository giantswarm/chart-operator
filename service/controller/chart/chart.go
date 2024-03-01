package chart

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/spf13/afero"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/chart-operator/v3/pkg/annotation"
	"github.com/giantswarm/chart-operator/v3/pkg/project"
	"github.com/giantswarm/chart-operator/v3/service/controller/chart/controllercontext"

	"github.com/giantswarm/chart-operator/v3/service/internal/clientpair"
)

const chartControllerSuffix = "-chart"

type Config struct {
	Fs          afero.Fs
	HelmClients *clientpair.ClientPair
	K8sClient   k8sclient.Interface
	Logger      micrologger.Logger

	ResyncPeriod time.Duration

	HTTPClientTimeout time.Duration
	K8sWaitTimeout    time.Duration
	K8sWatchNamespace string
	MaxRollback       int
	TillerNamespace   string
}

type Chart struct {
	*controller.Controller
}

func NewChart(config Config) (*Chart, error) {
	var err error

	if config.HelmClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClients must not be empty", config)
	}
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
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

	// TODO: Remove usage of deprecated controller context.
	//
	//	https://github.com/giantswarm/giantswarm/issues/12324
	//
	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		cc := controllercontext.Context{}
		ctx = controllercontext.NewContext(ctx, cc)

		return ctx, nil
	}

	var resources []resource.Interface
	{
		c := chartResourcesConfig{
			Fs:          config.Fs,
			CtrlClient:  config.K8sClient.CtrlClient(),
			HelmClients: config.HelmClients,
			K8sClient:   config.K8sClient.K8sClient(),
			Logger:      config.Logger,

			HTTPClientTimeout: config.HTTPClientTimeout,
			K8sWaitTimeout:    config.K8sWaitTimeout,
			MaxRollback:       config.MaxRollback,
			TillerNamespace:   config.TillerNamespace,
		}

		resources, err = newChartResources(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartController *controller.Controller
	{
		c := controller.Config{
			InitCtx:   initCtxFunc,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Pause: map[string]string{
				annotation.ChartOperatorPaused: "true",
			},
			Resources: resources,
			NewRuntimeObjectFunc: func() client.Object {
				return new(v1alpha1.Chart)
			},

			Name:      project.Name() + chartControllerSuffix,
			Namespace: config.K8sWatchNamespace,

			ResyncPeriod: config.ResyncPeriod,
		}

		chartController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Chart{
		Controller: chartController,
	}

	return c, nil
}
