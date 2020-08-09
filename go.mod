module github.com/giantswarm/chart-operator/v2

go 1.14

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/giantswarm/apiextensions/v2 v2.0.0-20200806181735-9cd3e758e49e
	github.com/giantswarm/appcatalog v0.2.7
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/helmclient/v2 v2.0.0-20200809132331-a75d70f49755
	github.com/giantswarm/k8sclient/v4 v4.0.0-20200806115259-2d3b230ace59
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit/v2 v2.0.0-20200809132745-a35924cf9504
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.5.1
	github.com/opencontainers/runc v1.0.0-rc2.0.20190611121236-6cc515888830 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/afero v1.3.3
	github.com/spf13/viper v1.7.0
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	sigs.k8s.io/yaml v1.2.0
)
