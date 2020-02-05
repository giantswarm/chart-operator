module github.com/giantswarm/chart-operator

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest v9.10.0+incompatible
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/goutils v1.1.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/PuerkitoBio/purell v1.1.1
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/beorn7/perks v1.0.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/chai2010/gettext-go v0.0.0-20191225085308-6b9f4b1008e1
	github.com/coreos/go-semver v0.3.0
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.11.1+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions v0.0.0-20200127104704-78e994410a10
	github.com/giantswarm/apprclient v0.0.0-20191209123802-955b7e89e6e2
	github.com/giantswarm/backoff v0.0.0-20190913091243-4dd491125192
	github.com/giantswarm/certctl v0.0.0-20200205145647-cd4e2d001215 // indirect
	github.com/giantswarm/e2e-harness v0.1.1-0.20191209145255-b2048d8645c1
	github.com/giantswarm/e2esetup v0.0.0-20191209131007-01b9f9061692
	github.com/giantswarm/e2etemplates v0.0.0-20200205154352-f7663e1e66d9
	github.com/giantswarm/errors v0.0.0-20200205145547-c12c94d9a110
	github.com/giantswarm/exporterkit v0.0.0-20190619131829-9749deade60f
	github.com/giantswarm/gitrepo v0.0.0-20200205150303-d38c26ad12bb // indirect
	github.com/giantswarm/gscliauth v0.1.1-0.20200205154154-1f3a753c3d56 // indirect
	github.com/giantswarm/helmclient v0.0.0-20200205160527-61453b0b25de
	github.com/giantswarm/k8sclient v0.0.0-20191209120459-6cb127468cd6
	github.com/giantswarm/k8sportforward v0.0.0-20191209165148-21368288d82d
	github.com/giantswarm/microendpoint v0.0.0-20191121160659-e991deac2653
	github.com/giantswarm/microerror v0.1.1-0.20200205143715-01b76f66cae6
	github.com/giantswarm/microkit v0.0.0-20191023091504-429e22e73d3e
	github.com/giantswarm/micrologger v0.0.0-20200205144836-079154bcae45
	github.com/giantswarm/operatorkit v0.0.0-20200205163802-6b6e6b2c208b
	github.com/giantswarm/to v0.0.0-20191022113953-f2078541ec95
	github.com/giantswarm/valuemodifier v0.0.0-20200205145307-cb5c3d1aa74d // indirect
	github.com/giantswarm/versionbundle v0.0.0-20200205145509-6772c2bc7b34
	github.com/go-kit/kit v0.9.0
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/jsonpointer v0.19.3
	github.com/go-openapi/jsonreference v0.19.3
	github.com/go-openapi/spec v0.19.5
	github.com/go-openapi/swag v0.19.6
	github.com/go-stack/stack v1.8.0
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/protobuf v1.3.2
	github.com/google/btree v1.0.0
	github.com/google/go-cmp v0.4.0
	github.com/google/gofuzz v1.1.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1
	github.com/gorilla/mux v1.7.3
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/hcl v1.0.0
	github.com/huandu/xstrings v1.3.0
	github.com/imdario/mergo v0.3.8
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/json-iterator/go v1.1.9
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/lib/pq v1.3.0
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/magiconair/properties v1.8.1
	github.com/mailru/easyjson v0.7.0
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-wordwrap v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mitchellh/reflectwalk v1.0.1
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.6.0
	github.com/petar/GoLLRB v0.0.0-20190514000832-33fb24c13b99
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.0.0
	github.com/rubenv/sql-migrate v0.0.0-20200119084958-8794cecc920c
	github.com/russross/blackfriday v1.5.2
	github.com/spf13/afero v1.2.2
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/subosito/gotenv v1.2.0
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	gomodules.xyz/jsonpatch v2.0.1+incompatible
	google.golang.org/appengine v1.6.5
	google.golang.org/genproto v0.0.0-20200122232147-0452cf42e150
	google.golang.org/grpc v1.26.0
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7
	gopkg.in/gorp.v1 v1.7.2
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/ini.v1 v1.51.1
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/square/go-jose.v2 v2.4.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/apiserver v0.16.4
	k8s.io/cli-runtime v0.16.4
	k8s.io/client-go v0.16.4
	k8s.io/code-generator v0.16.6 // indirect
	k8s.io/component-base v0.16.4
	k8s.io/helm v2.16.1+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	k8s.io/kubectl v0.16.4
	k8s.io/kubernetes v1.16.6
	k8s.io/utils v0.0.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.1.0
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787
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
