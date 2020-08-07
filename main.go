package main

import (
	"context"
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/chart-operator/flag"
	"github.com/giantswarm/chart-operator/pkg/project"
	"github.com/giantswarm/chart-operator/server"
	"github.com/giantswarm/chart-operator/service"
)

var (
	f = flag.New()
)

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainWithError() error {
	var err error

	ctx := context.Background()

	// Create a new logger that is used by all packages.
	var newLogger micrologger.Logger
	{
		c := micrologger.Config{
			IOWriter: os.Stdout,
		}
		newLogger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Define server factory to create the custom server once all command line
	// flags are parsed and all microservice configuration is processed.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// New custom service implements the business logic.
		var newService *service.Service
		{
			c := service.Config{
				Logger: newLogger,

				Flag:  f,
				Viper: v,
			}
			newService, err = service.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Mask(err)))
			}

			go newService.Boot(ctx)
		}

		// New custom server that bundles microkit endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  newLogger,
				Service: newService,

				Viper: v,
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Mask(err)))
			}
		}

		return newServer
	}

	// Create a new microkit command that manages operator daemon.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        newLogger,
			ServerFactory: newServerFactory,

			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Helm.HTTP.ClientTimeout, "5s", "HTTP timeout for pulling chart tarballs.")
	daemonCommand.PersistentFlags().String(f.Service.Helm.Kubernetes.WaitTimeout, "10s", "Wait timeout when calling the Kubernetes API.")
	daemonCommand.PersistentFlags().Int(f.Service.Helm.MaxRollback, 3, "the maximum number of rollback attempts for pending apps.")
	daemonCommand.PersistentFlags().String(f.Service.Helm.TillerNamespace, "giantswarm", "Namespace for the Tiller pod.")
	daemonCommand.PersistentFlags().String(f.Service.Image.Registry, "quay.io", "Container image registry.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfig, "", "KubeConfig used to connect to Kubernetes. When empty other settings are used.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	err = newCommand.CobraCommand().Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
