package service

import (
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/chart-operator/service/controller/v1"
	"github.com/giantswarm/chart-operator/service/controller/v2"
)

func NewVersionBundles() []versionbundle.Bundle {
	var versionBundles []versionbundle.Bundle

	versionBundles = append(versionBundles, v1.VersionBundle())
	versionBundles = append(versionBundles, v2.VersionBundle())

	return versionBundles
}
