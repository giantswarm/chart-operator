package helm

import (
	"github.com/giantswarm/chart-operator/flag/service/helm/http"
	"github.com/giantswarm/chart-operator/flag/service/helm/kubernetes"
)

type Helm struct {
	HTTP            http.HTTP
	Kubernetes      kubernetes.Kubernetes
	MaxRollback     string
	TillerNamespace string
}
