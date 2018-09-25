package v4

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Added support for user configmaps.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "chart-operator",
		Version:    "0.4.0",
	}
}
