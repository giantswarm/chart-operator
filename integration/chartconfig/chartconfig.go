package chartconfig

import (
	"bytes"
	"html/template"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
)

type Config struct {
	ChartValuesConfig
}

type ChartValuesConfig struct {
	Channel                  string
	ConfigMapName            string
	ConfigMapNamespace       string
	ConfigMapResourceVersion string
	Name                     string
	Namespace                string
	Release                  string
	SecretName               string
	SecretNamespace          string
	SecretResourceVersion    string
	VersionBundleVersion     string
}

type ChartConfig struct {
	chartValuesConfig   ChartValuesConfig
	chartValuesTemplate *template.Template
}

func NewChartConfig(config Config) (*ChartConfig, error) {
	cc := &ChartConfig{
		chartValuesConfig:   config.ChartValuesConfig,
		chartValuesTemplate: template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues)),
	}
	return cc, nil
}

func (cc ChartConfig) ChartValuesConfig() ChartValuesConfig {
	return cc.chartValuesConfig
}

func (cc ChartConfig) ExecuteChartValuesTemplate() (string, error) {
	buf := &bytes.Buffer{}
	err := cc.chartValuesTemplate.Execute(buf, cc.chartValuesConfig)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (cc ChartConfig) SetChartValuesConfig(values ChartValuesConfig) ChartValuesConfig {
	cc.chartValuesConfig = values
	return cc.chartValuesConfig
}
