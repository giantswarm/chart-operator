package chart

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/chart-operator/pkg/project"
	"github.com/giantswarm/chart-operator/service/controller/chart/controllercontext"
)

const chartControllerSuffix = "-chart"

type Config struct {
	Fs         afero.Fs
	HelmClient helmclient.Interface
	K8sClient  k8sclient.Interface
	Logger     micrologger.Logger

	K8sWaitTimeout  time.Duration
	MaxRollback     int
	TillerNamespace string
}

type Chart struct {
	*controller.Controller
}

func NewChart(config Config) (*Chart, error) {
	var err error

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
			Fs:         config.Fs,
			G8sClient:  config.K8sClient.G8sClient(),
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,

			K8sWaitTimeout:  config.K8sWaitTimeout,
			MaxRollback:     config.MaxRollback,
			TillerNamespace: config.TillerNamespace,
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
			Resources: resources,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.Chart)
			},

			Name: project.Name() + chartControllerSuffix,
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
