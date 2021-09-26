/*
Copyright 2021, CTERA Networks

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"context"
	"strings"

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/ctera/kubefiler-operator/internal/conf"
)

const (
	instanceLabel = "app.kubernetes.io/instance"
)

func getDeployment(ctx context.Context, client client.Client, ns, name string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		deployment,
	)

	return deployment, err
}

func getOrCreateGatewayDeployment(ctx context.Context, client client.Client, cfg *conf.OperatorConfig, instance *kubefilerv1alpha1.KubeFiler, gatewaySecret *corev1.Secret) (*appsv1.Deployment, bool, error) {
	deploymentName := getGatewayDeploymentName(instance)

	// fetch the existing secret, if available
	deployment, err := getDeployment(ctx, client, instance.GetNamespace(), deploymentName)
	if err == nil {
		return deployment, false, nil
	}

	if errors.IsNotFound(err) {
		deployment, err = generateGatewayDeployment(deploymentName, cfg, instance, gatewaySecret.Name, client.Scheme())
		if err != nil {
			return nil, false, err
		}
		err = client.Create(ctx, deployment)
		if err != nil {
			return nil, false, err
		}
		// Deployment created successfully
		return deployment, true, nil
	}

	return nil, false, err
}

func getGatewayDeploymentName(instance *kubefilerv1alpha1.KubeFiler) string {
	return instance.GetName() + "-kubefiler"
}

func generateGatewayDeployment(name string, cfg *conf.OperatorConfig, instance *kubefilerv1alpha1.KubeFiler, gatewaySecretName string, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	deployment, err := buildGatewayDeploymentSpec(name, cfg, instance, gatewaySecretName)
	if err != nil {
		return nil, err
	}
	err = controllerutil.SetControllerReference(instance, deployment, scheme)
	return deployment, err
}

func buildGatewayDeploymentSpec(name string, cfg *conf.OperatorConfig, instance *kubefilerv1alpha1.KubeFiler, gatewaySecretName string) (*appsv1.Deployment, error) {
	// construct a deployment based on the following labels
	labels := labelsForKubeFiler(name)
	var size int32 = 1

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.GetNamespace(),
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotationsForKubeFiler(cfg.GatewayContainerName),
				},
				Spec: buildGatewayPodSpec(cfg, instance, gatewaySecretName),
			},
		},
	}
	return deployment, nil
}

// labelsForKubeFiler returns the labels for selecting the resources
// belonging to the given CR name.
func labelsForKubeFiler(name string) map[string]string {
	return map[string]string{
		// top level labes
		"app": "kubefiler",
		// k8s namespaced labels
		// See: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
		"app.kubernetes.io/name":       "kubefiler",
		instanceLabel:                  labelValue("kubefiler", name),
		"app.kubernetes.io/component":  "kubefiler",
		"app.kubernetes.io/part-of":    "kubefiler",
		"app.kubernetes.io/managed-by": "kubefiler-operator",
		// our namespaced labels
		"kubefiler-operator.ctera.com/service": labelValue(name),
	}
}

func labelValue(s ...string) string {
	out := strings.Join(s, "-")
	if len(out) > 63 {
		out = out[:63]
	}
	return out
}

func annotationsForKubeFiler(name string) map[string]string {
	return map[string]string{
		"kubectl.kubernetes.io/default-logs-container": name,
		"kubectl.kubernetes.io/default-container":      name,
	}
}
