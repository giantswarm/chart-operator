package chartconfig

import (
	"bytes"
	"html/template"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type Config struct {
	Logger micrologger.Logger

	*ChartValuesConfig
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
	logger micrologger.Logger

	chartValuesConfig ChartValuesConfig
}

func NewChartConfig(config Config) (*ChartConfig, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	cc := &ChartConfig{
		chartValuesConfig: *config.ChartValuesConfig,
	}
	return cc, nil
}

func (cc ChartConfig) ChartValuesConfig() *ChartValuesConfig {
	return &cc.chartValuesConfig
}

func (cc ChartConfig) ExecuteChartValuesTemplate() (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	cc.logger.Log(*cc.ChartValuesConfig())
	err := chartValuesTemplate.Execute(buf, cc.chartValuesConfig)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (cc ChartConfig) SetChartValuesConfig(values *ChartValuesConfig) *ChartValuesConfig {
	cc.chartValuesConfig = *values
	return &cc.chartValuesConfig
}
