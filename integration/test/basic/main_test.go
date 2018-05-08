// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/integration/setup"
)

var (
	f          *framework.Host
	helmClient *helmclient.Client
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var err error

	f, err = framework.NewHost(framework.HostConfig{})
	if err != nil {
		panic(err.Error())
	}

	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		panic(err.Error())
	}

	c := helmclient.Config{
		Logger:     l,
		K8sClient:  f.K8sClient(),
		RestConfig: f.RestConfig(),
	}
	helmClient, err = helmclient.New(c)
	if err != nil {
		panic(err.Error())
	}

	setup.WrapTestMain(f, helmClient, m)
}
