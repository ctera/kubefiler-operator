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

	kubefilerv1alpha1 "github.com/ctera/ctera-gateway-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// UsernameKey is the name of the key to read the username from when reading the secret
	GatewayPortalAddressKey = "address"
)

func getConfigMap(ctx context.Context, client client.Client, ns, name string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		configMap,
	)

	return configMap, err
}

func getOrCreateGatewayPortalConfigMap(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler, portal *kubefilerv1alpha1.KubeFilerPortal, ns string) (*corev1.ConfigMap, bool, error) {
	configMapName := getGatewayPortalConfigMapName(instance)

	// fetch the existing secret, if available
	configMap, err := getConfigMap(ctx, client, ns, configMapName)
	if err == nil {
		return configMap, false, nil
	}

	if errors.IsNotFound(err) {
		configMap, err = generateGatewayPortalConfigMap(client, instance, ns, configMapName, portal)
		if err != nil {
			return configMap, false, err
		}
		err = client.Create(ctx, configMap)
		if err != nil {
			return configMap, false, err
		}
		// Deployment created successfully
		return configMap, true, nil
	}

	return nil, false, err
}

func getGatewayPortalConfigMapName(instance *kubefilerv1alpha1.KubeFiler) string {
	return strings.Join(
		[]string{
			instance.GetNamespace(),
			instance.GetName(),
			"portal",
		},
		"-",
	)
}

func generateGatewayPortalConfigMap(client client.Client, instance *kubefilerv1alpha1.KubeFiler, ns, name string, portal *kubefilerv1alpha1.KubeFilerPortal) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Data: map[string]string{
			GatewayPortalAddressKey: portal.Spec.Address,
		},
	}

	// Set the instance as the owner of the secret
	err := controllerutil.SetControllerReference(instance, configMap, client.Scheme())

	return configMap, err
}
