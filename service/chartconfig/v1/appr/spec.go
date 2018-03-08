package appr

const (
	httpClientTimeout = 5
)

// Package represents a CNR application.
type Package struct {
	Release string `json:"release"`
}
