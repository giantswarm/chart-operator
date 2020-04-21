package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Deploy as a unique app in control plane clusters.",
				Kind:        versionbundle.KindAdded,
				URLs: []string{
					"https://github.com/giantswarm/chart-operator/pull/421",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       Name(),
		Version:    Version(),
	}
}
