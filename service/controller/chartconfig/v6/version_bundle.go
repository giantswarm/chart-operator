package v6

import (
	"github.com/giantswarm/versionbundle"
)

func VersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Add your changes here.",
				Kind:        versionbundle.KindAdded,
			},
		},
		Components: []versionbundle.Component{},
		Name:       "chart-operator",
		Version:    "0.6.0",
	}
}
