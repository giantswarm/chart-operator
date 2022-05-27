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
		Logger:             microloggertest.New(),
		NamespaceWhitelist: []string{"org-giantswarm"},
		PrvHelmClient:      prvHC,
		PubHelmClient:      pubHC,
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
		// this is the mode the chart operator runs in for
		// Workload Clusters and Management Clusters with older
		// chart operator versions.
		{
			name: "single client, customer app",
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
			name: "single client, giantswarm app",
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
		// this is the mode the chart operator runs in for the
		// Management Clusters with new chart operator version.
		{
			name: "split client, customer app",
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
			name: "split client, giantswarm app",
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
		{
			name: "split client, org-giantswarm app",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: "org-giantswarm",
						annotation.AppName:      "kyverno",
					},
				},
			},
			clientPair:     splitClient,
			expectedClient: prvHC,
		},
		{
			name: "split client, WC app operator",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: "demo0",
						annotation.AppName:      "app-operator-demo0",
					},
				},
				Spec: v1alpha1.ChartSpec{
					TarballURL: appOperatorChart + "-5.9.0.tgz",
				},
			},
			clientPair:     splitClient,
			expectedClient: prvHC,
		},
		{
			name: "split client, WC app operator modified",
			chart: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotation.AppNamespace: "demo0",
						annotation.AppName:      "app-operator-demo0",
					},
				},
				Spec: v1alpha1.ChartSpec{
					TarballURL: "https://demo.github.io/demo-catalog/app-operator-5.9.0.tgz",
				},
			},
			clientPair:     splitClient,
			expectedClient: pubHC,
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
