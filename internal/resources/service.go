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

	kubefilerv1alpha1 "github.com/ctera/kubefiler-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getService(ctx context.Context, client client.Client, ns, name string) (*corev1.Service, error) {
	service := &corev1.Service{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		service,
	)

	return service, err
}

func getOrCreateGatewayService(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler) (*corev1.Service, bool, error) {
	serviceName := getGatewayServiceName(instance)

	// fetch the existing secret, if available
	service, err := getService(ctx, client, instance.GetNamespace(), serviceName)
	if err == nil {
		return service, false, nil
	}

	if errors.IsNotFound(err) {
		service, err = generateGatewayService(client, instance, serviceName)
		if err != nil {
			return service, false, err
		}
		err = client.Create(ctx, service)
		if err != nil {
			return service, false, err
		}
		// Deployment created successfully
		return service, true, nil
	}

	return nil, false, err
}

func getGatewayServiceName(instance *kubefilerv1alpha1.KubeFiler) string {
	return instance.GetName() + "-kubefiler"
}

var svcSelectorKey = "kubefiler-operator.ctera.com/service"

func generateGatewayService(client client.Client, instance *kubefilerv1alpha1.KubeFiler, name string) (*corev1.Service, error) {
	// as of now we only generate ClusterIP type services
	labels := labelsForKubeFiler(name)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.GetNamespace(),
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				Name:     "mgmt",
				Protocol: corev1.ProtocolTCP,
				Port:     443,
			}},
			Selector: map[string]string{
				svcSelectorKey: labels[svcSelectorKey],
			},
		},
	}

	err := controllerutil.SetControllerReference(instance, service, client.Scheme())

	return service, err
}
