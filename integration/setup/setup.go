// +build k8srequired

package setup

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/teardown"
)

func WrapTestMain(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger, m *testing.M) {
	var v int
	var err error

	err = h.CreateNamespace("giantswarm")
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = helmClient.EnsureTillerInstalled(ctx)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if env.KeepResources() != "true" {
		// only do full teardown when not on CI
		if env.CircleCI() != "true" {
			err := teardown.Teardown(h, helmClient)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			h.Teardown()
		}
	}

	os.Exit(v)
}
