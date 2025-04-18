package helm

import (
	"github.com/giantswarm/chart-operator/v4/flag/service/helm/http"
	"github.com/giantswarm/chart-operator/v4/flag/service/helm/kubernetes"
)

type Helm struct {
	HTTP        http.HTTP
	Kubernetes  kubernetes.Kubernetes
	MaxRollback string

	// SplitClient determines usage of additional pubHelmClient impersonating
	// `default:automation` Service Account for App CRs created outside the
	// `giantswarm` namespace. When `false` Chart Operator runs under full
	// cluster admin permissions no matter the App CR namespace.
	SplitClient        string
	NamespaceWhitelist string

	TillerNamespace string
}
