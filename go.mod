module github.com/giantswarm/chart-operator

go 1.13

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/chai2010/gettext-go v0.0.0-20191225085308-6b9f4b1008e1 // indirect
	github.com/giantswarm/apiextensions v0.0.0-20200127104704-78e994410a10
	github.com/giantswarm/apprclient v0.0.0-20191209123802-955b7e89e6e2
	github.com/giantswarm/e2e-harness v0.1.1-0.20191209145255-b2048d8645c1
	github.com/giantswarm/e2esetup v0.0.0-20191209131007-01b9f9061692
	github.com/giantswarm/e2etemplates v0.0.0-20200205154352-f7663e1e66d9
	github.com/giantswarm/exporterkit v0.0.0-20190619131829-9749deade60f
	github.com/giantswarm/helmclient v0.0.0-20200205160527-61453b0b25de
	github.com/giantswarm/k8sclient v0.0.0-20191209120459-6cb127468cd6
	github.com/giantswarm/k8sportforward v0.0.0-20191209165148-21368288d82d
	github.com/giantswarm/microendpoint v0.0.0-20200205204116-c2c5b3af4bdb
	github.com/giantswarm/microerror v0.1.1-0.20200205143715-01b76f66cae6
	github.com/giantswarm/microkit v0.0.0-20191023091504-429e22e73d3e
	github.com/giantswarm/micrologger v0.0.0-20200205144836-079154bcae45
	github.com/giantswarm/operatorkit v0.0.0-20200205163802-6b6e6b2c208b
	github.com/giantswarm/versionbundle v0.0.0-20200205145509-6772c2bc7b34
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.4.0
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/viper v1.6.2
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82 // indirect
	google.golang.org/genproto v0.0.0-20200122232147-0452cf42e150 // indirect
	google.golang.org/grpc v1.26.0 // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.16.4
	k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go v0.16.4
	k8s.io/helm v2.16.1+incompatible
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.1+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	// All of that is because helm has an import to k8s.io/kubernetes which
	// uses relative paths to those.
	k8s.io/api v0.0.0 => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.16.4
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.16.4
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.16.4
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.16.4
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.16.4
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.16.4
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.16.4
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.16.4
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.16.4
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.16.4
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.16.4
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.16.4
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.16.4
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.16.4
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.16.4
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.16.4
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.16.4
	k8s.io/utils v0.0.0 => k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6
)
