// +build k8srequired

package integration

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	helmclient "k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
)

const (
	chartOperatorValues = ``
)

var (
	f *framework.Host
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var v int
	var err error
	f, err = framework.NewHost(framework.HostConfig{})
	if err != nil {
		log.Printf("unexpected error: %v\n", err)
		os.Exit(1)
	}

	if err := f.CreateNamespace("giantswarm"); err != nil {
		log.Printf("unexpected error: %v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		f.Teardown()
	}

	os.Exit(v)
}

func TestGetReleaseVersion(t *testing.T) {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Errorf("could not create logger %v", err)
	}

	c := appr.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      "http://localhost:5000",
		Organization: "giantswarm",
	}

	a, err := appr.New(c)
	if err != nil {
		t.Errorf("could not create appr %v", err)
	}

	expected := "3.2.1"
	actual, err := a.GetReleaseVersion("test-chart", "3-2-beta")
	if err != nil {
		t.Errorf("could not get release %v", err)
	}

	if expected != actual {
		t.Errorf("release didn't match expected, want %q, got %q", expected, actual)
	}
}

func TestInstallChart(t *testing.T) {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
	if err != nil {
		t.Fatalf("could not create k8s config %v", err)
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("could not create k8s client %v", err)
	}

	hc := helm.Config{
		Logger:     l,
		K8sClient:  cs,
		RestConfig: config,
	}

	h, err := helm.New(hc)
	if err != nil {
		t.Fatalf("could not create helm client %v", err)
	}

	// --test-dir dir is mounted in /e2e in the test container.
	tarballPath := filepath.Join("/e2e", "tb-chart.tar.gz")

	const releaseName = "tb-chart-release"

	err = h.InstallFromTarball(tarballPath, "default", helmclient.ReleaseName(releaseName))
	if err != nil {
		t.Fatalf("could not install chart %v", err)
	}

	releaseContent, err := h.GetReleaseContent(releaseName)
	if err != nil {
		t.Fatalf("could not get release content %v", err)
	}

	expectedName := releaseName
	actualName := releaseContent.Name
	if expectedName != actualName {
		t.Fatalf("bad release name, want %q, got %q", expectedName, actualName)
	}

	expectedStatus := "DEPLOYED"
	actualStatus := releaseContent.Status
	if expectedStatus != actualStatus {
		t.Fatalf("bad release status, want %q, got %q", expectedStatus, actualStatus)
	}

	err = h.DeleteRelease(releaseName)
	if err != nil {
		t.Fatalf("could not delete release %v", err)
	}

	releaseContent, err = h.GetReleaseContent(releaseName)
	if err != nil {
		t.Fatalf("could not get release content %v", err)
	}
	expectedStatus = "DELETED"
	actualStatus = releaseContent.Status
	if expectedStatus != actualStatus {
		t.Fatalf("bad release status, want %q, got %q", expectedStatus, actualStatus)
	}
}

func TestInstallOperator(t *testing.T) {
	err := f.InstallOperator("chart-operator", "chartconfig", chartOperatorValues, ":${CIRCLE_SHA1}")

	if err != nil {
		t.Fatalf("could not install operator %v", err)
	}
}
