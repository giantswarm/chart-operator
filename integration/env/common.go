package env

import (
	"os"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"
)

var (
	circleCI      string
	keepResources string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)
}

func CircleCI() string {
	return circleCI
}

func KeepResources() string {
	return keepResources
}
