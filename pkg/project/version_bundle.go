package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "Only update failed Helm releases if the chart values or version has changed.",
				Kind:        versionbundle.KindAdded,
				URLs: []string{
					"https://github.com/giantswarm/chart-operator/pull/422",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       Name(),
		Version:    Version(),
	}
}
