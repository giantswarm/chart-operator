package chartmigration

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	chartConfigNotInstalledErrorText = "the server could not find the requested resource (get chartconfigs.core.giantswarm.io)"
)

var (
	chartConfigNotAvailablePatterns = []*regexp.Regexp{
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/apis/core.giantswarm.io/v1alpha1/namespaces.* (unexpected )?EOF`),
		regexp.MustCompile(`[Get|Patch|Post] https://api\..*/apis/core.giantswarm.io/v1alpha1/namespaces.* net/http: (TLS handshake timeout|request canceled).*?`),
	}
)

var chartConfigNotAvailableError = &microerror.Error{
	Kind: "ChartConfigNotAvailableError",
}

func isChartConfigNotAvailable(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	for _, re := range chartConfigNotAvailablePatterns {
		matched := re.MatchString(c.Error())

		if matched {
			return true
		}
	}

	return c == chartConfigNotAvailableError
}

var chartConfigNotInstalledError = &microerror.Error{
	Kind: "ChartConfigNotInstalledError",
}

func isChartConfigNotInstalled(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.Contains(c.Error(), chartConfigNotInstalledErrorText) {
		return true
	}

	if c == chartConfigNotInstalledError {
		return true
	}

	return false
}
