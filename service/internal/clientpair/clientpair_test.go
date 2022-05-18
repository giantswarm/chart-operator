package clientpair

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/helmclient/v4/pkg/helmclienttest"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/micrologger/microloggertest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_NewClientPair(t *testing.T) {
	testCases := []struct {
		name         string
		config       ClientPairConfig
		errorMatcher func(error) bool
	}{
		{
			name: "flawless, single client",
			config: ClientPairConfig{
				Logger:        microloggertest.New(),
				PrvHelmClient: helmclienttest.New(helmclienttest.Config{}),
				PubHelmClient: nil,
			},
		},
		{
			name: "flawless, split client",
			config: ClientPairConfig{
				Logger:        microloggertest.New(),
				PrvHelmClient: helmclienttest.New(helmclienttest.Config{}),
				PubHelmClient: helmclienttest.New(helmclienttest.Config{}),
			},
		},
		{
			name: "missing private client",
			config: ClientPairConfig{
				Logger:        microloggertest.New(),
				PrvHelmClient: nil,
				PubHelmClient: helmclienttest.New(helmclienttest.Config{}),
			},
			errorMatcher: IsInvalidConfig,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			_, err := NewClientPair(tc.config)

			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case tc.errorMatcher != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}
		})
	}
}

func Test_Get(t *testing.T) {
	prvHC := helmclienttest.New(helmclienttest.Config{})
	pubHC := helmclienttest.New(helmclienttest.Config{})

	singleClient, err := NewClientPair(ClientPairConfig{
		Logger:        microloggertest.New(),
		PrvHelmClient: prvHC,
		PubHelmClient: nil,
	})
	if err != nil {
		t.Fatalf("error == %#v, want nil", err)
	}

	splitClient, err := NewClientPair(ClientPairConfig{
		Logger:        microloggertest.New(),
		PrvHelmClient: prvHC,
		PubHelmClient: pubHC,
	})
	if err != nil {
		t.Fatalf("error == %#v, want nil", err)
	}

	testCases := []struct {
		name           string
		chart          v1alpha1.Chart
		clientPair     *ClientPair
		expectedClient helmclient.Interface
	}{
		{
			name: "flawless, single client, outside giantswarm ns",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: "org-acme",
						annotation.AppName:      "test",
					},
				},
			},
			clientPair:     singleClient,
			expectedClient: prvHC,
		},
		{
			name: "flawless, split client, outside giantswarm ns",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: "org-acme",
						annotation.AppName:      "test",
					},
				},
			},
			clientPair:     splitClient,
			expectedClient: pubHC,
		},
		{
			name: "flawless, single client, giantswarm ns",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: privateNamespace,
						annotation.AppName:      "kyverno",
					},
				},
			},
			clientPair:     singleClient,
			expectedClient: prvHC,
		},
		{
			name: "flawless, split client, giantswarm ns",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: privateNamespace,
						annotation.AppName:      "kyverno",
					},
				},
			},
			clientPair:     splitClient,
			expectedClient: prvHC,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			client := tc.clientPair.Get(context.TODO(), tc.chart)

			if client != tc.expectedClient {
				t.Fatalf("got wrong client")
			}
		})
	}
}
