package appr

import "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

const (
	httpClientTimeout = 5
)

// Interface describes the methods provided by the appr client.
type Interface interface {
	GetReleaseVersion(v1alpha1.ChartConfig) (string, error)
}

// Channel represents a CNR channel.
type Channel struct {
	Current string `json:"current"`
}
