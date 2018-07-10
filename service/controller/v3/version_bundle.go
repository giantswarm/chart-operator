package v3

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Added support for setting the chartconfig CR status with the Helm release status.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "chart-operator",
		Version:    "0.3.0",
	}
}
