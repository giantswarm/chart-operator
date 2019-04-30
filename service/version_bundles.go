package service

import (
	"github.com/giantswarm/versionbundle"

	chartv1 "github.com/giantswarm/chart-operator/service/controller/chart/v1"

	chartconfigv1 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v1"
	chartconfigv2 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v2"
	chartconfigv3 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v3"
	chartconfigv4 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v4"
	chartconfigv5 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v5"
	chartconfigv6 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v6"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, chartv1.VersionBundle())

	versionBundles = append(versionBundles, chartconfigv1.VersionBundle())
	versionBundles = append(versionBundles, chartconfigv2.VersionBundle())
	versionBundles = append(versionBundles, chartconfigv3.VersionBundle())
	versionBundles = append(versionBundles, chartconfigv4.VersionBundle())
	versionBundles = append(versionBundles, chartconfigv5.VersionBundle())
	versionBundles = append(versionBundles, chartconfigv6.VersionBundle())

	return versionBundles
}
