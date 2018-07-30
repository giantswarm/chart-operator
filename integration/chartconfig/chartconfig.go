// +build k8srequired

package chartconfig

import (
	"bytes"
	"html/template"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
)

type ChartConfigValues struct {
	Channel              string
	ConfigMap            ConfigMap
	Name                 string
	Namespace            string
	Release              string
	Secret               Secret
	VersionBundleVersion string
}

type ConfigMap struct {
	Name            string
	Namespace       string
	ResourceVersion string
}

type Secret struct {
	Name            string
	Namespace       string
	ResourceVersion string
}

func (ccv ChartConfigValues) ExecuteChartValuesTemplate() (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
