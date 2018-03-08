package appr

import "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

const (
	httpClientTimeout = 5
)

// Interface describes the methods provided by the appr client.
type Interface interface {
	GetRelease(v1alpha1.ChartConfig) (string, error)
}

// Package represents a CNR application.
type Package struct {
	Release string `json:"release"`
}
