package service

import (
	"github.com/giantswarm/versionbundle"

	chartv1 "github.com/giantswarm/chart-operator/service/controller/chart/v1"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, chartv1.VersionBundle())

	return versionBundles
}
