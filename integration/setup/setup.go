// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/giantswarm/chart-operator/integration/env"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	var v int
	var err error

	err = config.Host.CreateNamespace("giantswarm")
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "failed to create namespace", "stack", fmt.Sprintf("%#v", err))
		v = 1
	}

	err = config.HelmClient.EnsureTillerInstalled(ctx)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "failed to install tiller", "stack", fmt.Sprintf("%#v", err))
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if env.KeepResources() != "true" {
		// only do full teardown when not on CI
		if env.CircleCI() != "true" {
			err := teardown(ctx, config)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			config.Host.Teardown()
		}
	}

	os.Exit(v)
}
