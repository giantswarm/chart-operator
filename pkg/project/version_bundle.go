package project

import (
	"github.com/giantswarm/versionbundle"
)

func NewVersionBundle() versionbundle.Bundle {
	return versionbundle.Bundle{
		Changelogs: []versionbundle.Changelog{
			{
				Component:   "chart-operator",
				Description: "TODO",
				Kind:        versionbundle.KindChanged,
				URLs: []string{
					"",
				},
			},
		},
		Components: []versionbundle.Component{},
		Name:       Name(),
		Version:    Version(),
	}
}
