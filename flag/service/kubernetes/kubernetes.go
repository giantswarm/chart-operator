package kubernetes

import (
	"github.com/giantswarm/chart-operator/flag/service/kubernetes/tls"
)

// Kubernetes is a data structure to hold Kubernetes specific command line
// configuration flags.
type Kubernetes struct {
	Address   string
	InCluster string
	TLS       tls.TLS
}
