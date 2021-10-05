//go:build k8srequired
// +build k8srequired

package key

func DefaultCatalogStorageURL() string {
	return "https://giantswarm.github.io/default-catalog"
}

func DefaultTestCatalogStorageURL() string {
	return "https://giantswarm.github.io/default-test-catalog"
}
