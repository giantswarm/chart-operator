package v7

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Check cordon annotations in chart resource to avoid any further update.",
				Kind:        versionbundle.KindAdded,
			},
			{
				Component:   "chart-operator",
				Description: "Add delete annotation support for chartconfig to app CR migration.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "chart-operator",
		Version:    "0.7.0",
	}
}
