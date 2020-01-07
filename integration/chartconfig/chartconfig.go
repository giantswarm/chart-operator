// +build k8srequired

package chartconfig

import (
	"bytes"
	"context"
	"html/template"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/setup"
)

func DeleteResources(ctx context.Context, config setup.Config) error {
	items := []string{"cnr-server", "giantswarm-apiextensions-chart-config-e2e-chart"}

	for _, item := range items {
		err := config.HelmClient.DeleteRelease(ctx, item, helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func ExecuteValuesTemplate(ccv e2etemplates.ApiextensionsChartConfigValues) (string, error) {
	buf := &bytes.Buffer{}
	chartValuesTemplate := template.Must(template.New("chartConfigChartValues").Parse(e2etemplates.ApiextensionsChartConfigE2EChartValues))
	err := chartValuesTemplate.Execute(buf, ccv)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return buf.String(), nil
}

func InstallResources(ctx context.Context, config setup.Config) error {
	err := initializeCNR(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func initializeCNR(ctx context.Context, config setup.Config) error {
	err := installCNR(ctx, config)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCNR(ctx context.Context, config setup.Config) error {
	var err error

	{
		replicas := int32(1)
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cnr-server",
				Namespace: metav1.NamespaceDefault,
				Labels: map[string]string{
					"app": "cnr-server",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "cnr-server",
					},
				},
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cnr-server",
						Labels: map[string]string{
							"app": "cnr-server",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "cnr-server",
								Image:           "quay.io/giantswarm/cnr-server:latest",
								ImagePullPolicy: corev1.PullIfNotPresent,
							},
						},
					},
				},
			},
		}

		_, err = config.K8sClients.K8sClient().AppsV1().Deployments(metav1.NamespaceDefault).Create(deployment)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		service := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cnr-server",
				Namespace: metav1.NamespaceDefault,
				Labels: map[string]string{
					"app": "cnr-server",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "cnr-server",
						Port:     int32(5000),
						Protocol: "TCP",
					},
				},
				Selector: map[string]string{
					"app": "cnr-server",
				},
			},
		}

		_, err = config.K8sClients.K8sClient().CoreV1().Services(metav1.NamespaceDefault).Create(service)
		if apierrors.IsAlreadyExists(err) {
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
