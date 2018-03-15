package appr_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
)

func Test_GetReleaseVersion(t *testing.T) {
	customObject := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chartname",
				Channel: "3-2-beta",
			},
		},
	}
	tcs := []struct {
		description     string
		h               func(w http.ResponseWriter, r *http.Request)
		expectedError   bool
		expectedRelease string
	}{
		{
			description: "basic case",
			h: func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/cnr/api/v1/packages/giantswarm/chartname/channels/3-2-beta") {
					http.Error(w, "wrong path", http.StatusInternalServerError)
					fmt.Println(r.URL.Path)
					return
				}

				c := appr.Channel{
					Current: "3.2.1",
				}
				js, err := json.Marshal(c)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
			},
			expectedRelease: "3.2.1",
		},
		{
			description: "wrongly formated response",
			h: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("not json"))
			},
			expectedError: true,
		},
		{
			description: "server error",
			h: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "500!!", http.StatusInternalServerError)
				return
			},
			expectedError: true,
		},
	}

	c := appr.Config{
		Logger:       microloggertest.New(),
		Organization: "giantswarm",
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tc.h))
			defer ts.Close()

			c.Address = ts.URL
			a, err := appr.New(c)
			if err != nil {
				t.Errorf("could not create appr %v", err)
			}

			r, err := a.GetReleaseVersion(customObject)
			switch {
			case err != nil && !tc.expectedError:
				t.Errorf("failed to get release %v", err)
			case err == nil && tc.expectedError:
				t.Errorf("expected error didn't happen")
			}

			if r != tc.expectedRelease {
				t.Errorf("didn't get expected release name, want %q, got %q", tc.expectedRelease, r)
			}
		})
	}
}

func Test_DeleteRelease(t *testing.T) {
	customObject := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chartname",
				Channel: "3-2-beta",
			},
		},
	}
	tcs := []struct {
		description     string
		h               func(w http.ResponseWriter, r *http.Request)
		expectedError   bool
		expectedRelease string
	}{
		{
			description: "basic case",
			h: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					msg := fmt.Sprintf("wrong method %q, expected DLETE", r.Method)
					http.Error(w, msg, http.StatusInternalServerError)
					return
				}
				if !strings.HasPrefix(r.URL.Path, "/cnr/api/v1/packages/giantswarm/chartname/channels/3-2-beta") {
					http.Error(w, "wrong path", http.StatusInternalServerError)
					fmt.Println(r.URL.Path)
					return
				}

				c := appr.Channel{
					Current: "3.2.1",
				}
				js, err := json.Marshal(c)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
			},
			expectedRelease: "3.2.1",
		},
		{
			description: "wrongly formated response",
			h: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("not json"))
			},
			expectedError: true,
		},
		{
			description: "server error",
			h: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "500!!", http.StatusInternalServerError)
				return
			},
			expectedError: true,
		},
	}

	c := appr.Config{
		Logger:       microloggertest.New(),
		Organization: "giantswarm",
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tc.h))
			defer ts.Close()

			c.Address = ts.URL
			a, err := appr.New(c)
			if err != nil {
				t.Errorf("could not create appr %v", err)
			}

			err = a.DeleteRelease(customObject)
			switch {
			case err != nil && !tc.expectedError:
				t.Errorf("failed to get release %v", err)
			case err == nil && tc.expectedError:
				t.Errorf("expected error didn't happen")
			}
		})
	}
}
