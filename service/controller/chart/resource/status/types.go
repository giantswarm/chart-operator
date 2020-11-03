package status

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Request struct {
	AppVersion   string  `json:"app_version"`
	LastDeployed v1.Time `json:"last_deployed"`
	Reason       string  `json:"reason"`
	Status       string  `json:"status"`
	Token        string  `json:"token"`
	Version      string  `json:"version"`
}
