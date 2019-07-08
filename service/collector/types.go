package collector

type chartState struct {
	chartName     string
	channelName   string
	cordonUntil   float64
	namespace     string
	releaseName   string
	releaseStatus string
}
