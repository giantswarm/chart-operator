package service

import (
	"github.com/giantswarm/operatorkit/v7/pkg/flag/service/kubernetes"

	"github.com/giantswarm/chart-operator/v3/flag/service/helm"
	"github.com/giantswarm/chart-operator/v3/flag/service/image"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Helm       helm.Helm
	Image      image.Image
	Kubernetes kubernetes.Kubernetes
}
