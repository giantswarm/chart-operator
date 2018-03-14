package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/microkit/transaction"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
	"github.com/spf13/viper"

	"github.com/giantswarm/chart-operator/flag"
	"github.com/giantswarm/chart-operator/server"
	"github.com/giantswarm/chart-operator/service"
)

const (
	notAvailable = "n/a"
)

var (
	description = "The chart-operator deploys Helm charts by reconciling against a CNR repository."
	f           = flag.New()
	name        = "chart-operator"
	gitCommit   = notAvailable
	source      = "https://github.com/giantswarm/chart-operator"
)

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", err))
	}
}

func mainWithError() (err error) {
	// Create a new logger that is used by all packages.
	var newLogger micrologger.Logger
	{
		c := micrologger.Config{
			IOWriter: os.Stdout,
		}
		newLogger, err = micrologger.New(c)
		if err != nil {
			return microerror.Maskf(err, "micrologger.New")
		}
	}

	// Define server factory to create the custom server once all command line
	// flags are parsed and all microservice configuration is processed.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// New custom service implements the business logic.
		var newService *service.Service
		{
			c := service.Config{
				Flag:   f,
				Logger: newLogger,
				Viper:  v,

				Description: description,
				GitCommit:   notAvailable,
				ProjectName: name,
				Source:      source,
			}
			newService, err = service.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "service.New")))
			}

			go newService.Boot()
		}

		var storage microstorage.Storage
		{
			storage, err = memory.New(memory.DefaultConfig())
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "memory.New")))
			}
		}

		var transactionResponder transaction.Responder
		{
			c := transaction.DefaultResponderConfig()
			c.Logger = newLogger
			c.Storage = storage

			transactionResponder, err = transaction.NewResponder(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "transaction.NewResponder")))
			}
		}

		// New custom server that bundles microkit endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				MicroServerConfig: microserver.DefaultConfig(),
				Service:           newService,
			}

			c.MicroServerConfig.Logger = newLogger
			c.MicroServerConfig.ServiceName = name
			c.MicroServerConfig.TransactionResponder = transactionResponder
			c.MicroServerConfig.Viper = v

			newServer, err = server.New(c)
			if err != nil {
				panic(fmt.Sprintf("%#v\n", microerror.Maskf(err, "server.New")))
			}
		}

		return newServer
	}

	// Create a new microkit command that manages operator daemon.
	var newCommand command.Command
	{
		commandConfig := command.DefaultConfig()

		commandConfig.Logger = newLogger
		commandConfig.ServerFactory = newServerFactory

		commandConfig.Description = description
		commandConfig.GitCommit = notAvailable
		commandConfig.Name = name
		commandConfig.Source = source

		newCommand, err = command.New(commandConfig)
		if err != nil {
			return microerror.Maskf(err, "command.New")
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.CNR.Address, "https://quay.io", "Address used to connect to CNR, defaults to quay's managed offering.")
	daemonCommand.PersistentFlags().String(f.Service.CNR.Organization, "giantswarm", "CNR organization.")
	daemonCommand.PersistentFlags().String(f.Service.Helm.Host, ":44134", "Host and port for Helm client to connect to the Tiller endpoint.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	newCommand.CobraCommand().Execute()

	return nil
}
