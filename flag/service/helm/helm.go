package helm

import (
	"github.com/giantswarm/chart-operator/flag/service/helm/http"
)

type Helm struct {
	HTTP            http.HTTP
	TillerImageName string
	TillerNamespace string
}
