package service

import (
	"github.com/giantswarm/versionbundle"

	chartv1 "github.com/giantswarm/chart-operator/service/controller/chart/v1"

	v5 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v5"
	v6 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v6"
	v7 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v7"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, chartv1.VersionBundle())
§
versionBundles = append(versionBundles, chartconfigv5.VersionBundle())
versionBundles = append(versionBundles, chartconfigv6.VersionBundle())
versionBundles = append(versionBundles, chartconfigv7.VersionBundle())

	return versionBundles
}
