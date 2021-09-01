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

func getServiceAccount(ctx context.Context, client client.Client, ns, name string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		serviceAccount,
	)

	return serviceAccount, err
}

func getOrCreateServiceAccount(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler) (*corev1.ServiceAccount, bool, error) {
	serviceAccountName := getServiceAccountName(instance)
	serviceAccount, err := getServiceAccount(ctx, client, instance.GetNamespace(), serviceAccountName)
	if err == nil {
		return serviceAccount, false, nil
	}

	if errors.IsNotFound(err) {
		serviceAccount, err = generateServiceAccount(client, instance, serviceAccountName)
		if err != nil {
			return serviceAccount, false, err
		}

		err = client.Create(ctx, serviceAccount)
		if err != nil {
			return serviceAccount, false, err
		}
		// RoleBinding created successfully
		return serviceAccount, true, nil
	}
	return nil, false, err
}

func getServiceAccountName(instance *kubefilerv1alpha1.KubeFiler) string {
	return instance.GetName() + "-filer"
}

func generateServiceAccount(client client.Client, instance *kubefilerv1alpha1.KubeFiler, name string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.GetNamespace(),
		},
	}

	// Set the instance as the owner of the secret
	err := controllerutil.SetControllerReference(instance, serviceAccount, client.Scheme())

	return serviceAccount, err
}
