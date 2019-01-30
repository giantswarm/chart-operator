// +build k8srequired

package chartconfig

import (
	"bytes"
	"html/template"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
)

func ExecuteChartConfigValuesTemplate(ccv e2etemplates.ApiextensionsChartConfigValues) (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
