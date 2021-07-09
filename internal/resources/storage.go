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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubefilerv1alpha1 "github.com/ctera/ctera-gateway-operator/api/v1alpha1"
)

func kubeFilerNeedsPvc(instance *kubefilerv1alpha1.KubeFiler) bool {
	return instance.Spec.Storage.Pvc != nil && instance.Spec.Storage.Pvc.Spec != nil
}

func getPvc(ctx context.Context, client client.Client, ns, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := client.Get(
		ctx,
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
		pvc,
	)

	return pvc, err
}

func getOrCreateGatewayPvc(ctx context.Context, client client.Client, instance *kubefilerv1alpha1.KubeFiler, ns string) (*corev1.PersistentVolumeClaim, bool, error) {
	pvcName := getGatewayPvcName(instance)

	// fetch the existing secret, if available
	pvc, err := getPvc(ctx, client, ns, pvcName)
	if err == nil {
		return pvc, false, nil
	}

	if errors.IsNotFound(err) {
		pvc, err = generateGatewayPvc(client, instance, ns, pvcName)
		if err != nil {
			return pvc, false, err
		}
		err = client.Create(ctx, pvc)
		if err != nil {
			return pvc, false, err
		}
		// Deployment created successfully
		return pvc, true, nil
	}

	return nil, false, err
}

func getGatewayPvcName(instance *kubefilerv1alpha1.KubeFiler) string {
	if instance.Spec.Storage.Pvc.Name != "" {
		return instance.Spec.Storage.Pvc.Name
	}
	return instance.GetName() + "-pvc"
}

func generateGatewayPvc(client client.Client, instance *kubefilerv1alpha1.KubeFiler, ns, name string) (*corev1.PersistentVolumeClaim, error) {
	// build a new pvc
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: *instance.Spec.Storage.Pvc.Spec,
	}

	// set the kube filer instance as the owner and controller
	err := controllerutil.SetControllerReference(instance, pvc, client.Scheme())

	return pvc, err

}
