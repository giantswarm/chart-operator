package release

import (
	"context"

	"github.com/giantswarm/helmclient"
	"k8s.io/helm/pkg/helm"
)

type helmMock struct {
	defaultReleaseContent *helmclient.ReleaseContent
	defaultReleaseHistory *helmclient.ReleaseHistory
	defaultError          error
}

func (h *helmMock) DeleteRelease(ctx context.Context, releaseName string, options ...helm.DeleteOption) error {
	if h.defaultError != nil {
		return h.defaultError
	}

	return nil
}

func (h *helmMock) EnsureTillerInstalled(ctx context.Context) error {
	return nil
}

func (h *helmMock) GetReleaseContent(ctx context.Context, releaseName string) (*helmclient.ReleaseContent, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseContent, nil
}

func (h *helmMock) GetReleaseHistory(ctx context.Context, releaseName string) (*helmclient.ReleaseHistory, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseHistory, nil
}

func (h *helmMock) InstallReleaseFromTarball(ctx context.Context, path, ns string, options ...helm.InstallOption) error {
	return nil
}

func (h *helmMock) ListReleaseContents(ctx context.Context) ([]*helmclient.ReleaseContent, error) {
	return nil, nil
}

func (h *helmMock) RunReleaseTest(ctx context.Context, releaseName string, options ...helm.ReleaseTestOption) error {
	return nil
}

func (h *helmMock) UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}

func (h *helmMock) PingTiller(ctx context.Context) error {
	return nil
}
