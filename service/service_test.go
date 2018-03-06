package service

import (
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/viper"

	"github.com/giantswarm/chart-operator/flag"
)

func Test_Service_New(t *testing.T) {
	testCases := []struct {
		name         string
		config       func() Config
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: valid config returns no error",
			config: func() Config {
				c := Config{
					Flag:   flag.New(),
					Logger: microloggertest.New(),
					Viper:  viper.New(),

					ProjectName: "chart-operator",
				}

				c.Viper.Set(c.Flag.Service.Kubernetes.Address, "https://127.0.0.1:8443")
				c.Viper.Set(c.Flag.Service.CNR.Address, "https://127.0.0.1:5555")
				c.Viper.Set(c.Flag.Service.Kubernetes.InCluster, false)

				return c
			},
			errorMatcher: nil,
		},
		{
			name: "case 1: invalid config returns error",
			config: func() Config {
				c := Config{
					Flag:  flag.New(),
					Viper: viper.New(),
				}

				return c
			},
			errorMatcher: IsInvalidConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.config())

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
