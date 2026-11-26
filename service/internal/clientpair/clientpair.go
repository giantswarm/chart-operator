package clientpair

import (
	"context"
	"strings"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/v4/service/controller/chart/key"
)

const (
	// privateNamespace defines GS-protected namespace the prvHelmClient
	// is meant for. For App CRs created outside this namespace, the
	// pubHelmClient should be used.
	privateNamespace = "giantswarm"

	// WC App Operators are the only cases where the elevated prvHelmClient
	// has to be used outside the `giantswarm` namespace. It is due to these
	// apps installing reources the `default:automation` does not have access to.
	// However, existance of the App CR for unique App Operator outside the
	// `giantswarm` namespace opens a door to scenarios described here:
	// https://github.com/giantswarm/giantswarm/issues/22100.
	// The appOperatorChart defines legal prefix for what elevated Helm Client
	// can install outside the `giantswarm` namespace.
	appOperatorChart = "https://giantswarm.github.io/control-plane-catalog/app-operator"
	// The appOperatorTestChart defines another legal prefix to use the elevated Helm client
	// used when we are testing changes for workload cluster app-operators that are
	// generally not located in the giantswarm namespace.
	appOperatorTestChart = "https://giantswarm.github.io/control-plane-test-catalog/app-operator"

	// The tarball URL alone does not prove an app operator chart comes from
	// Giant Swarm: a tenant-owned Catalog can list a chart under any name with
	// its URL pointing at the control-plane tarball. The elevated client is
	// only used when app-operator resolved the chart from the matching Giant
	// Swarm catalog, which tenants cannot create or change. See
	// https://github.com/giantswarm/giantswarm/issues/39470.
	appOperatorCatalog     = "control-plane-catalog"
	appOperatorTestCatalog = "control-plane-test-catalog"
)

// giantswarmCatalogNamespaces are the namespaces the Giant Swarm catalogs live
// in. app-operator resolves an App without catalogNamespace from these.
var giantswarmCatalogNamespaces = map[string]bool{
	"default":    true,
	"giantswarm": true,
}

type ClientPairConfig struct {
	Logger micrologger.Logger

	NamespaceWhitelist []string

	PrvHelmClient helmclient.Interface
	PubHelmClient helmclient.Interface
}

type ClientPair struct {
	logger micrologger.Logger

	namespaceWhitelist []string

	prvHelmClient helmclient.Interface
	pubHelmClient helmclient.Interface
}

func NewClientPair(config ClientPairConfig) (*ClientPair, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.PrvHelmClient == helmclient.Interface(nil) {
		return nil, microerror.Maskf(invalidConfigError, "%T.PrvHelmClient must not be empty", config)
	}

	cp := &ClientPair{
		logger: config.Logger,

		namespaceWhitelist: config.NamespaceWhitelist,

		prvHelmClient: config.PrvHelmClient,
		pubHelmClient: config.PubHelmClient,
	}

	return cp, nil
}

// Get determines which client to use based on the namespace the corresponding App CR
// is located in. For Workload Cluster, chart operator is permitted to operate under
// cluster-wide permissions, so there is only prvHelmClient used.
func (cp *ClientPair) Get(ctx context.Context, cr v1alpha1.Chart, privateClient bool) helmclient.Interface {
	// nil pubHelmClient means chart-operator runs in a single-client mode
	// under cluster admin privileges.
	if cp.pubHelmClient == helmclient.Interface(nil) {
		return cp.prvHelmClient
	}

	// for App CRs created inside the `giantswarm` namespace, use the prvHelmClient
	// that runs under cluster admin privileges.
	if key.AppNamespace(cr) == privateNamespace {
		cp.logger.Debugf(ctx, "selecting private Helm client for `%s` App in `%s` namespace", key.AppName(cr), key.AppNamespace(cr))

		return cp.prvHelmClient
	}

	// extra check against additional whitelisted namespaces. The privateNamespace
	// is hardcoded because it is well-known namespace.
	for _, ns := range cp.namespaceWhitelist {
		if key.AppNamespace(cr) == ns {
			cp.logger.Debugf(ctx, "selecting private Helm client for `%s` App in `%s` namespace", key.AppName(cr), key.AppNamespace(cr))

			return cp.prvHelmClient
		}
	}

	// for app operators outside the `giantswarm` namespace, use the prvHelmClient.
	if key.AppNamespace(cr) != privateNamespace && isGiantSwarmAppOperator(cr) {
		cp.logger.Debugf(ctx, "selecting private Helm client for `%s` App in `%s` namespace", key.AppName(cr), key.AppNamespace(cr))

		return cp.prvHelmClient
	}

	// select private Helm client when requested. Use it with caution. It has been
	// introduce to answer permissions issue when deleting Chart CRs,
	// see: https://github.com/giantswarm/giantswarm/issues/25731
	if privateClient {
		cp.logger.Debugf(ctx, "selecting private Helm client for `%s` App in `%s` namespace on demand", key.AppName(cr), key.AppNamespace(cr))

		return cp.prvHelmClient
	}

	// for App CRs created outside the `giantswarm` namespace, or not carrying the
	// annotation in question, use the pubHelmClient that runs under `automation`
	// Service Account privileges.
	cp.logger.Debugf(ctx, "selecting public Helm client for `%s` App in `%s` namespace", key.AppName(cr), key.AppNamespace(cr))

	return cp.pubHelmClient
}

// isGiantSwarmAppOperator reports whether the chart is the app operator chart
// from a Giant Swarm catalog: its tarball URL is in the control-plane (test)
// catalog and app-operator resolved it from that catalog in a Giant Swarm
// namespace.
func isGiantSwarmAppOperator(cr v1alpha1.Chart) bool {
	if !giantswarmCatalogNamespaces[key.CatalogNamespace(cr)] {
		return false
	}

	switch key.CatalogName(cr) {
	case appOperatorCatalog:
		return strings.HasPrefix(key.TarballURL(cr), appOperatorChart)
	case appOperatorTestCatalog:
		return strings.HasPrefix(key.TarballURL(cr), appOperatorTestChart)
	default:
		return false
	}
}
